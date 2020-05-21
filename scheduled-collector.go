package main

import (
	"context"
	"encoding/json"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

func setIfSuccess(dst *string, value string) {
	if value != "" {
		*dst = value
	}
}

// some
type scheduledCollector struct {
	option CollectorOptions
	// props
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// k8s
	PodName     string
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
	Version string
	Commit  string

	// metrics
	NodeInfo prometheus.Gauge
	// TODO: persistence-check
	UDStateSize    prometheus.Gauge
	UDStateEntries prometheus.Gauge
}

func NewScheduledCollector(option CollectorOptions) *scheduledCollector {
	c := &scheduledCollector{option: option}
	return c
}

func (s *scheduledCollector) Init(register prometheus.Registerer) {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.GetStatics()
	s.NodeInfo = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_info",
		Help: "Node Information of zilliqa and host environment",
		ConstLabels: prometheus.Labels{
			"pod_name": s.PodName, "namespace": s.Namespace, "node_name": s.NodeName, "network_name": s.NetworkName,
			"placement": s.Placement, "cluster_name": s.ClusterName,
			"public_ip": s.PublicIP, "public_hostname": s.PublicHostname,
			"local_ip": s.LocalIP, "local_hostname": s.LocalHostname,
			"instance_id": s.InstanceID, "instance_type": s.InstanceType,
		},
	})
	register.MustRegister(s.NodeInfo)

	if s.option.IsMainNet {
		s.UDStateSize = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ud_state_size",
			Help: "State data size of unstoppable domain contract",
			ConstLabels: prometheus.Labels{
				"cluster_name": s.ClusterName, "network_name": s.NetworkName, "pod_name": s.PodName, "public_ip": s.PublicIP, "local_ip": s.LocalIP,
			},
		})
		s.UDStateEntries = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ud_state_entries",
			Help: "State entries count of unstoppable domain contract",
			ConstLabels: prometheus.Labels{
				"cluster_name": s.ClusterName, "network_name": s.NetworkName, "pod_name": s.PodName, "public_ip": s.PublicIP, "local_ip": s.LocalIP,
			},
		})
		register.MustRegister(s.UDStateSize, s.UDStateEntries)
	}
}

func (s *scheduledCollector) GetStatics() {
	s.PodName = getEnvKeys("POD_NAME", "Z7A_POD_NAME")
	s.NodeName = getEnvKeys("NAMESPACE")
	s.NodeName = getEnvKeys("NODE_NAME", "Z7A_NODE_NAME")
	s.NetworkName = getEnvKeys("Z7A_TESTNET_NAME", "TESTNET_NAME", "NETWORK_NAME")
	s.Commit = getEnvKeys("ZILLIQA_COMMIT")

	s.ClusterName = getEnvKeys("CLUSTER_NAME")

	s.Placement = getMetadata("placement/availability-zone")

	s.PublicIP = getMetadata("public-ipv4")
	s.PublicHostname = getMetadata("public-hostname")

	s.LocalIP = getMetadata("local-ipv4")
	s.LocalHostname = getMetadata("local-hostname")

	s.InstanceID = getMetadata("instance-id")
	s.InstanceType = getMetadata("instance-type")

	if s.option.IsSameNS {
		s.Version = getExecOutput("zilliqa", "-v")
	}
}

func (s *scheduledCollector) GetUDContractStateSizeRecords() (int, int, error) {
	var address = "9611c53BE6d1b32058b2747bdeCECed7e1216793"
	api := provider.NewProvider(s.option.APIEndpoint)
	resp, err := api.GetSmartContractState(address)
	if err != nil {
		return 0, 0, err
	}
	js, err := json.Marshal(resp.Result)
	if err != nil {
		return 0, 0, err
	}
	size := len(js)
	var result map[string]json.RawMessage
	err = json.Unmarshal(js, &result)
	if err != nil {
		return 0, 0, err
	}
	records, _ := result["records"]
	var entries map[string]interface{}
	err = json.Unmarshal(records, &entries)
	if err != nil {
		return 0, 0, err
	}
	return size, len(entries), nil
}

func (s *scheduledCollector) CollectGetUDContractStateSizeRecords() {
	size, records, err := s.GetUDContractStateSizeRecords()
	if err != nil {
		log.WithError(err).Error("fail to get UD Contract state")
	}
	s.UDStateSize.Set(float64(size))
	s.UDStateEntries.Set(float64(records))
}

type ScheduleTask struct {
	fun      func()
	interval time.Duration
}

func (s *scheduledCollector) Start() {

	schedules := []ScheduleTask{
		{s.GetStatics, 30 * time.Minute},
	}

	if s.option.IsMainNet {
		schedules = append(schedules, ScheduleTask{s.CollectGetUDContractStateSizeRecords, 1 * time.Hour})
	}

	for _, t := range schedules {
		go func() {
			s.wg.Add(1)
			defer s.wg.Done()
			for {
				select {
				case <-time.After(t.interval):
					t.fun()
				case <-s.ctx.Done():
					return
				}
			}
		}()
	}
}

func (s *scheduledCollector) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
