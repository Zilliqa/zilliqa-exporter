package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/zilliqa/zilliqa-exporter/utils"
	"strconv"
	"strings"
	"time"
)

type Constants struct {
	options *Options
	// props
	//ctx    context.Context
	//cancel context.CancelFunc
	//wg     sync.WaitGroup

	// k8s
	PodName     string
	PodIP       string
	Namespace   string
	NodeName    string
	NetworkName string

	// aws & testnet & genet
	Placement      string
	ClusterName    string
	PublicIP       string
	PublicHostname string
	LocalIP        string
	LocalHostname  string
	InstanceID     string
	InstanceType   string

	// zilliqa binary
	BinPath string
	Version string
	Commit  string

	// Desc
	NodeInfo *prometheus.Desc

	// detected
	detectStop chan struct{}
	nodeType   NodeType
	nodeIndex  int
	p2pPort    uint32
}

func NewConstants(options *Options) *Constants {
	c := &Constants{options: options}
	c.NodeInfo = prometheus.NewDesc(
		"node_info",
		"Node Information of zilliqa and host environment",
		[]string{
			"pod_name", "short_pod_name", "pod_ip", "namespace", "node_name", "network_name", "cluster_name", // network related
			"placement", "instance_id", "instance_type", // ec2 instance related
			"public_ip", "public_hostname",
			"local_ip", "local_hostname",
			"zilliqa_bin_path", "zilliqa_version", // fs related
			"type", "index", // process related and may change or detected after exporter starts
		},
		nil,
	)
	c.doCollect()
	c.doDetectVars()
	return c
}

func (c *Constants) ShortPodName() string {
	split := strings.Split(c.PodName, "-") // xxx-TYPE-INDEX (generated pod name of stateful set)
	if len(split) > 2 {
		return fmt.Sprintf("%s-%s", split[len(split)-2], split[len(split)-1])
	}
	return c.PodName
}

func (c *Constants) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.NodeInfo
}

func (c *Constants) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.NodeInfo, prometheus.GaugeValue, 1,
		c.PodName, c.ShortPodName(), c.PodIP, c.Namespace, c.NodeName, c.NetworkName, c.ClusterName,
		c.Placement, c.InstanceID, c.InstanceType,
		c.PublicIP, c.PublicHostname,
		c.LocalIP, c.LocalHostname,
		c.BinPath, c.Version,
		c.NodeType().String(), strconv.Itoa(c.nodeIndex),
	)
}

func (c *Constants) doCollect() {
	// from envVars
	c.PodName = utils.GetEnvKeys("POD_NAME", "Z7A_POD_NAME")
	c.PodIP = utils.GetEnvKeys("POD_IP", "Z7A_POD_IP")
	c.Namespace = utils.GetEnvKeys("NAMESPACE")
	c.NodeName = utils.GetEnvKeys("NODE_NAME", "Z7A_NODE_NAME")
	c.NetworkName = utils.GetEnvKeys("Z7A_TESTNET_NAME", "TESTNET_NAME", "NETWORK_NAME")
	c.Commit = utils.GetEnvKeys("ZILLIQA_COMMIT")

	c.ClusterName = utils.GetEnvKeys("CLUSTER_NAME")

	// from AWS metadata
	if utils.MetadataAvailable() {
		c.Placement = utils.GetMetadata("placement/availability-zone")

		c.PublicIP = utils.GetMetadata("public-ipv4")
		c.PublicHostname = utils.GetMetadata("public-hostname")

		c.LocalIP = utils.GetMetadata("local-ipv4")
		c.LocalHostname = utils.GetMetadata("local-hostname")

		c.InstanceID = utils.GetMetadata("instance-id")
		c.InstanceType = utils.GetMetadata("instance-type")
	} else {
		log.Error("AWS Metadata not available")
	}

	// from file system
	if bin := c.options.ZilliqaBinPath(); bin != "" {
		c.BinPath = bin
		c.Version = strings.TrimSpace(utils.GetExecOutput(bin, "-v"))
	}
}

func (c *Constants) CommonLabels() []string {
	return []string{
		"type", "index",
		"cluster_name", "network_name", "pod_name", "pod_ip", "public_ip", "local_ip",
	}
}

func (c *Constants) CommonLabelValues() []string {
	return []string{
		c.NodeType().String(), strconv.Itoa(c.nodeIndex),
		c.ClusterName, c.NetworkName, c.PodName, c.PodIP, c.PublicIP, c.LocalIP,
	}
}

func (c *Constants) NodeType() NodeType {
	return c.nodeType
}

func (c *Constants) P2PPort() uint32 {
	return c.p2pPort
}

func nodeTypeIndexFromPodName(podName string) (NodeType, int) {
	split := strings.Split(podName, "-") // xxx-TYPE-INDEX (generated pod name of stateful set)
	if len(split) > 2 {
		nt := NodeTypeFromString(split[len(split)-2])
		index, err := strconv.Atoi(split[len(split)-1])
		if err != nil {
			return nt, -1
		}
		return nt, index
	}
	return UnknownNodeType, -1
}

func (c *Constants) doDetectVars() {
	var nodeTypeDetected bool
	var p2pPortDetected bool

	c.nodeType = NodeTypeFromString(utils.GetEnvKeys("NODE_TYPE"))
	c.nodeIndex, _ = strconv.Atoi(utils.GetEnvKeys("NODE_TYPE"))

	if c.PodName != "" {
		nt, idx := nodeTypeIndexFromPodName(c.PodName)
		if nt != UnknownNodeType {
			c.nodeType = nt
			c.nodeIndex = idx
			nodeTypeDetected = true
		}
	}

	var cmdline []string
	var err error
	if pd := GetZilliqadProcess(); pd != nil {
		cmdline, err = pd.CmdlineSlice()
	} else if p := GetZilliqaMainProcess(c); p != nil {
		cmdline, err = p.CmdlineSlice()
	}
	if err != nil {
		log.WithError(err).Error("fail to get cmdline")
		return
	} else {
		if nt := GetNodeTypeFromCmdline(cmdline); !nodeTypeDetected && nt != "" {
			c.nodeType = NodeTypeFromString(nt)
			nodeTypeDetected = true
		}

		if idx, err := GetNodeIndexFromCmdline(cmdline); err == nil {
			c.nodeIndex = idx
		}

		if p2p, err := GetPortFromCmdline(cmdline); err == nil {
			c.p2pPort = uint32(p2p)
			p2pPortDetected = true
		}
	}

	if !p2pPortDetected {
		c.p2pPort = c.options.p2pPort
		log.Debug("unable to auto-detect p2p port")
	}
	if !nodeTypeDetected {
		c.nodeType = NodeTypeFromString(c.options.nodeType)
		log.Debug("unable to auto-detect node type")
	}
}

func (c *Constants) StartWatch() {
	log.Info("start watching constants")
	if c.detectStop != nil {
		close(c.detectStop)
	}
	c.detectStop = make(chan struct{})
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.detectStop:
			log.Debug("stop schedule collecting constants")
			return
		case <-ticker.C:
			log.Debug("schedule collecting constants")
			c.doCollect()
		}
	}
}

func (c *Constants) StopWatch() {
	if c.detectStop != nil {
		close(c.detectStop)
	}
	log.Info("stop watching constants")
}
