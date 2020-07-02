package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/zilliqa/zilliqa-exporter/utils"
	"strings"
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
	nodeInfo prometheus.Gauge
}

func GetConstants(options *Options) *Constants {
	c := &Constants{options: options}
	c.Init()
	c.DetectNodeType()
	return c
}

func (c *Constants) Init() {
	c.PodName = utils.GetEnvKeys("POD_NAME", "Z7A_POD_NAME")
	c.PodIP = utils.GetEnvKeys("POD_IP", "Z7A_POD_IP")
	c.Namespace = utils.GetEnvKeys("NAMESPACE")
	c.NodeName = utils.GetEnvKeys("NODE_NAME", "Z7A_NODE_NAME")
	c.NetworkName = utils.GetEnvKeys("Z7A_TESTNET_NAME", "TESTNET_NAME", "NETWORK_NAME")
	c.Commit = utils.GetEnvKeys("ZILLIQA_COMMIT")

	c.ClusterName = utils.GetEnvKeys("CLUSTER_NAME")

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

	if bin := c.options.ZilliqaBinPath(); bin != "" {
		c.BinPath = bin
		c.Version = strings.TrimSpace(utils.GetExecOutput(bin, "-v"))
	}
}

func (c *Constants) ConstLabels() prometheus.Labels {
	return prometheus.Labels{
		"type":         c.options.NodeType().String(),
		"cluster_name": c.ClusterName,
		"network_name": c.NetworkName,
		"pod_name":     c.PodName,
		"pod_ip":       c.PodIP,
		"public_ip":    c.PublicIP,
		"local_ip":     c.LocalIP,
	}
}

func (c *Constants) Register(registerer prometheus.Registerer) {
	c.nodeInfo = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_info",
		Help: "Node Information of zilliqa and host environment",
		ConstLabels: prometheus.Labels{
			"pod_name": c.PodName, "pod_ip": c.PodIP, "type": c.options.NodeType().String(),
			"namespace": c.Namespace, "node_name": c.NodeName, "network_name": c.NetworkName,
			"placement": c.Placement, "cluster_name": c.ClusterName,
			"public_ip": c.PublicIP, "public_hostname": c.PublicHostname,
			"local_ip": c.LocalIP, "local_hostname": c.LocalHostname,
			"instance_id": c.InstanceID, "instance_type": c.InstanceType,
			"zilliqa_bin_path": c.BinPath, "zilliqa_version": c.Version,
		},
	})
	c.nodeInfo.Set(1)
	registerer.MustRegister(c.nodeInfo)
}

// map nodeType returned by admin client to nodeType of options
//var adminNodeTypeNameToNodeTypeMap = map[adminclient.NodeTypeName]string {
//	adminclient.ShardNode:
//}

func (c *Constants) DetectNodeType() {
	if c.PodName != "" {
		split := strings.Split(c.PodName, "-") // xxx-TYPE-INDEX (generated pod name of stateful set)
		if len(split) > 2 {
			name := NodeTypeFromString(split[len(split)-2])
			if string(name) != "" {
				DetectedNodeType = name
				return
			}
		}
	}
	if p := utils.GetZilliqaMainProcess(); p != nil {
		var cmdline []string
		var err error
		pd := utils.GetZilliqadProcess()
		if pd != nil {
			cmdline, err = pd.CmdlineSlice()
		} else {
			cmdline, err = p.CmdlineSlice()
		}
		if err == nil {
			return
		}
		nt := GetNodeTypeFromCmdline(cmdline)
		if nt != "" {
			DetectedNodeType = NodeTypeFromString(nt)
			return
		}
	}
	// TODO: get node type from admin api or log
	//adminEp := c.options.AdminEndpoint()
	//if adminEp != "" {
	//	cli := adminclient.New(adminEp, 3*time.Second)
	//	nt, err := cli.GetNodeType()
	//	if err == nil{
	//		nt.Type
	//	}
	//}

	log.Debug("unable to auto-detect node type")
}
