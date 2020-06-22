package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

// collect instant values
type Instant struct {
	option Options

	// block-chain related
	// from admin & api
	epoch   *prometheus.Desc
	dsEpoch *prometheus.Desc

	difficulty   *prometheus.Desc
	dsDifficulty *prometheus.Desc

	// misc node information, from both admin and lookup api
	// GetNetworkId of api
	networkID *prometheus.Desc
	// GetNodeType of admin
	//   NotJoined
	//   Seed
	//   Lookup
	//   DSNode
	//   ShardNode
	nodeType *prometheus.Desc
	shardNum *prometheus.Desc

	// enum DirState : unsigned char {
	//	POW_SUBMISSION = 0x00,
	//	DSBLOCK_CONSENSUS_PREP,
	//	DSBLOCK_CONSENSUS,
	//	MICROBLOCK_SUBMISSION,
	//	FINALBLOCK_CONSENSUS_PREP,
	//	FINALBLOCK_CONSENSUS,
	//	VIEWCHANGE_CONSENSUS_PREP,
	//	VIEWCHANGE_CONSENSUS,
	//	ERROR,
	//	SYNC
	// }
	// Not to be queried on lookup
	nodeState *prometheus.Desc

	// from lookup api
	//
}

func NewInstantCollector(option Options, constLabels prometheus.Labels) *Instant {
	return &Instant{
		option: option,
		epoch: prometheus.NewDesc(
			"epoch", "Current TX block number of the node",
			nil, constLabels,
		),
		dsEpoch: prometheus.NewDesc(
			"ds_epoch", "Current DS block number of the node",
			nil, constLabels,
		),
		difficulty: prometheus.NewDesc(
			"difficulty", "The minimum shard difficulty of the previous block",
			nil, constLabels,
		),
		dsDifficulty: prometheus.NewDesc(
			"ds_difficulty", "The minimum DS difficulty of the previous block",
			nil, constLabels,
		),
		networkID: prometheus.NewDesc(
			"network_id", "Network ID of current zilliqa network",
			nil, constLabels,
		),
		nodeType: prometheus.NewDesc(
			"node_type", "Zilliqa network node type",
			nil, constLabels,
		),
		shardNum: prometheus.NewDesc(
			"shard_num", "Shard number of current node",
			nil, constLabels,
		),
		nodeState: prometheus.NewDesc(
			"node_state", "Node state",
			nil, constLabels,
		),
	}
}

func (c *Instant) Describe(ch chan<- *prometheus.Desc) {
	// api & admin
	//if c.options.APIEndpoint != "" || c.options.AdminEndpoint() != "" {
	//	ch <- c.epoch
	//	ch <- c.dsEpoch
	//
	//	ch <- c.difficulty
	//	ch <- c.dsDifficulty
	//}
	//
	//if c.options.APIEndpoint != "" {
	//	ch <- c.networkID
	//}
	//
	//// api
	//if c.options.AdminAPIEndpoint != "" || c.options.IsSideCar {
	//	ch <- c.nodeType
	//	ch <- c.shardId
	//	ch <- c.nodeState
	//}

}

func (c *Instant) Collect(ch chan<- prometheus.Metric) {
	//wg := sync.WaitGroup{}
	//var hasAPI bool
	//if c.options.CheckEndpoint() == nil {
	//	hasAPI = true
	//	cli := c.options.GetAPIClient()
	//	wg.Add(1)
	//	go func() {
	//		log.Debug("start getting BlockchainInfo")
	//		defer wg.Done()
	//		resp, err := cli.GetBlockchainInfo()
	//		if err != nil {
	//			log.WithError(err).Error("fail to get block chain info")
	//		}
	//		info := main.BlockchainInfo{}
	//		err = resp.GetObject(&info)
	//		if err != nil {
	//			log.WithError(err).Error("fail to parse block chain info")
	//		}
	//		epoch, _ := strconv.Atoi(info.CurrentMiniEpoch)
	//		ch <- prometheus.MustNewConstMetric(c.epoch, prometheus.GaugeValue, float64(epoch))
	//		dsEpoch, _ := strconv.Atoi(info.CurrentDSEpoch)
	//		ch <- prometheus.MustNewConstMetric(c.dsEpoch, prometheus.GaugeValue, float64(dsEpoch))
	//		//ch <- prometheus.MustNewConstMetric(c.difficulty,  prometheus.GaugeValue, float64(info.))
	//		log.Debug("exit getting BlockchainInfo")
	//	}()
	//
	//	wg.Add(1)
	//	go func() {
	//		log.Debug("start getting GetPrevDifficulty")
	//		defer wg.Done()
	//		resp, err := cli.GetPrevDifficulty()
	//		if err != nil {
	//			log.WithError(err).Error("fail to get block prev difficulty")
	//		}
	//		difficulty, err := resp.GetFloat()
	//		if err != nil {
	//			log.WithError(err).Error("fail to parse block chain prev difficulty")
	//		}
	//		ch <- prometheus.MustNewConstMetric(c.difficulty, prometheus.GaugeValue, difficulty)
	//		log.Debug("exit getting GetPrevDifficulty")
	//	}()
	//
	//	wg.Add(1)
	//	go func() {
	//		log.Debug("start getting GetPrevDSDifficulty")
	//		defer wg.Done()
	//		resp, err := cli.GetPrevDSDifficulty()
	//		if err != nil {
	//			log.WithError(err).Error("fail to get block prev DS difficulty")
	//		}
	//		dsDifficulty, err := resp.GetFloat()
	//		if err != nil {
	//			log.WithError(err).Error("fail to parse block chain prev DS difficulty")
	//		}
	//		ch <- prometheus.MustNewConstMetric(c.dsDifficulty, prometheus.GaugeValue, dsDifficulty)
	//		log.Debug("exit getting GetPrevDSDifficulty")
	//	}()
	//
	//	wg.Add(1)
	//	go func() {
	//		log.Debug("start getting GetNetworkId")
	//		defer wg.Done()
	//		resp, err := cli.GetNetworkId()
	//		if err != nil {
	//			log.WithError(err).Error("fail to get network id")
	//		}
	//		netID, err := resp.GetInt()
	//		if err != nil {
	//			log.WithError(err).Error("fail to parse network id")
	//		}
	//		ch <- prometheus.MustNewConstMetric(c.networkID, prometheus.GaugeValue, float64(netID))
	//		log.Debug("exit getting GetNetworkId")
	//	}()
	//}
	//
	////if c.options.CheckGetAdminClient() == nil {
	////	//ch <- prometheus.MustNewConstMetric(c.epoch, prometheus.GaugeValue, 0)
	////	//ch <- prometheus.MustNewConstMetric(c.dsEpoch, prometheus.GaugeValue, 0)
	////}
	//if hasAPI {
	//}
	////wg.Wait()
	time.Sleep(3 * time.Second)
	log.Debug("exit instant collector")
}
