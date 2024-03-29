package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/zilliqa/zilliqa-exporter/adminclient"
	"github.com/zilliqa/zilliqa-exporter/jsonrpc"
)

// collect instant values
type AdminCollector struct {
	options   *Options
	constants *Constants

	// admin api server up
	adminServerUp *prometheus.Desc

	// block-chain related
	// from admin & api
	epoch   *prometheus.Desc
	dsEpoch *prometheus.Desc

	difficulty   *prometheus.Desc
	dsDifficulty *prometheus.Desc
	// GetNodeType of admin
	//   NotJoined
	//   Seed
	//   Lookup
	//   DSNode
	//   ShardNode
	nodeType *prometheus.Desc
	shardId  *prometheus.Desc

	// https://github.com/Zilliqa/Zilliqa/blob/master/src/libDirectoryService/DirectoryService.h
	// enum DirState : unsigned char {
	//	POW_SUBMISSION = 0x00,
	//	DSBLOCK_CONSENSUS_PREP = 1,
	//	DSBLOCK_CONSENSUS = 2,
	//	MICROBLOCK_SUBMISSION = 3,
	//	FINALBLOCK_CONSENSUS_PREP = 4,
	//	FINALBLOCK_CONSENSUS = 5,
	//	VIEWCHANGE_CONSENSUS_PREP = 6,
	//	VIEWCHANGE_CONSENSUS = 7,
	//	ERROR = 8,
	//	SYNC = 9
	// }
	// Not to be queried on lookup
	nodeState *prometheus.Desc
}

func NewAdminCollector(constants *Constants) *AdminCollector {
	commonLabels := constants.CommonLabels()
	return &AdminCollector{
		options:   constants.options,
		constants: constants,
		adminServerUp: prometheus.NewDesc(
			"admin_server_up", "Admin JsonRPC server (status server) up and running",
			append([]string{"endpoint"}, commonLabels...), nil,
		),
		epoch: prometheus.NewDesc(
			"epoch", "Current TX block number of the node",
			commonLabels, nil,
		),
		dsEpoch: prometheus.NewDesc(
			"ds_epoch", "Current DS block number of the node",
			commonLabels, nil,
		),
		difficulty: prometheus.NewDesc(
			"difficulty", "The minimum shard difficulty of the previous block",
			commonLabels, nil,
		),
		dsDifficulty: prometheus.NewDesc(
			"ds_difficulty", "The minimum DS difficulty of the previous block",
			commonLabels, nil,
		),
		nodeType: prometheus.NewDesc(
			"node_type", "Zilliqa network node type",
			append([]string{"text"}, commonLabels...), nil,
		),
		shardId: prometheus.NewDesc(
			"shard_id", "Shard ID of the shard of current node",
			commonLabels, nil,
		),
		nodeState: prometheus.NewDesc(
			"node_state", "Node state",
			append([]string{"text"}, commonLabels...), nil,
		),
	}
}

func (c *AdminCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.adminServerUp
	ch <- c.nodeType
	ch <- c.epoch
	ch <- c.dsEpoch
	ch <- c.difficulty
	ch <- c.dsDifficulty
	if !IsGeneralLookup(c.constants.NodeType()) {
		ch <- c.shardId
		//ch <- c.nodeState
	}
}

func (c *AdminCollector) Collect(ch chan<- prometheus.Metric) {
	labels := c.constants.CommonLabelValues()
	log.Debug("enter admin collector")
	cli := c.options.GetAdminClient()
	if cli == nil {
		log.Error("API endpoint not set")
		return
	}
	log.Debug("GetNodeType from admin API")
	nodeType, err := cli.GetNodeType()
	if err != nil {
		log.WithError(err).Error("error while getting NodeType from admin API")
		ch <- prometheus.MustNewConstMetric(c.adminServerUp, prometheus.GaugeValue, float64(0),
			append([]string{c.options.AdminEndpoint()}, labels...)...)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.adminServerUp, prometheus.GaugeValue, float64(1),
		append([]string{c.options.AdminEndpoint()}, labels...)...)
	ch <- prometheus.MustNewConstMetric(c.nodeType, prometheus.GaugeValue, float64(nodeType.Type),
		append([]string{nodeType.Type.String()}, labels...)...)

	if nodeType.Type == adminclient.ShardNode {
		ch <- prometheus.MustNewConstMetric(c.shardId, prometheus.GaugeValue, float64(nodeType.ShardId), labels...)
	}

	reqs := []*jsonrpc.Request{
		adminclient.NewGetCurrentMiniEpochReq(),
		adminclient.NewGetCurrentDSEpochReq(),
		adminclient.NewGetPrevDifficultyReq(),
		adminclient.NewGetPrevDSDifficultyReq(),
		//adminclient.NewGetNodeStateReq(),
	}

	log.Debug("batch GetCurrentMiniEpoch, GetCurrentDSEpoch, GetPrevDifficulty, GetPrevDSDifficulty, GetNodeState from admin API")
	resps, err := cli.CallBatch(reqs...)
	if err != nil {
		log.WithError(err).Error("error while getting non-lookup infos from admin API")
	} else if len(resps) != len(reqs) {
		log.Errorf("unknown error while getting non-lookup infos from admin API, responses less than %d", len(reqs))
	}

	epoch, err := resps[0].GetFloat64()
	if err != nil {
		log.WithError(err).Error("error while getting miniEpoch from admin API")
	} else {
		ch <- prometheus.MustNewConstMetric(c.epoch, prometheus.GaugeValue, epoch, labels...)
	}

	dsEpoch, err := resps[1].GetFloat64()
	if err != nil {
		log.WithError(err).Error("error while getting dsEpoch from admin API")
	} else {
		ch <- prometheus.MustNewConstMetric(c.dsEpoch, prometheus.GaugeValue, dsEpoch, labels...)
	}

	diff, err := resps[2].GetFloat64()
	if err != nil {
		log.WithError(err).Error("error while getting prevDifficulty from admin API")
	} else {
		ch <- prometheus.MustNewConstMetric(c.difficulty, prometheus.GaugeValue, diff, labels...)
	}

	dsDiff, err := resps[3].GetFloat64()
	if err != nil {
		log.WithError(err).Error("error while getting prevDSDifficulty from admin API")
	} else {
		ch <- prometheus.MustNewConstMetric(c.dsDifficulty, prometheus.GaugeValue, dsDiff, labels...)
	}

	// TODO: node state
	//var state adminclient.NodeState
	//err = resps[4].GetObject(&state)
	//if err != nil {
	//	log.WithError(err).Error("error while getting nodeState from admin API")
	//} else {
	//	ch <- prometheus.MustNewConstMetric(c.nodeState, prometheus.GaugeValue, float64(state), state.String(), labels...)
	//}
	log.Debug("exit admin collector")
}
