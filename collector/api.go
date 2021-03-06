package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/zilliqa/zilliqa-exporter/utils"
	"strconv"
	"sync"
	"time"
)

// collect instant values
type APICollector struct {
	options   *Options
	constants *Constants

	// is jsonrpc api server up
	apiServerUp *prometheus.Desc

	// block-chain related
	// from GetBlockchainInfo
	epoch           *prometheus.Desc
	dsEpoch         *prometheus.Desc
	transactionRate *prometheus.Desc
	txBlockRate     *prometheus.Desc
	dsBlockRate     *prometheus.Desc
	numPeers        *prometheus.Desc
	numTransactions *prometheus.Desc
	numTxBlocks     *prometheus.Desc
	numDSBlocks     *prometheus.Desc
	numTxnsDSEpoch  *prometheus.Desc
	numTxnsTxEpoch  *prometheus.Desc
	// sharding structure
	shardingPeers *prometheus.Desc

	difficulty   *prometheus.Desc
	dsDifficulty *prometheus.Desc

	// misc
	networkID              *prometheus.Desc
	latestTxBlockTimestamp *prometheus.Desc
	latestDsBlockTimestamp *prometheus.Desc

	apiServerDetected bool
}

func NewAPICollector(constants *Constants) *APICollector {
	commonLabels := constants.CommonLabels()
	return &APICollector{
		options:   constants.options,
		constants: constants,
		apiServerUp: prometheus.NewDesc(
			"api_server_up", "JsonRPC API server up and running",
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
		transactionRate: prometheus.NewDesc(
			"transaction_rate", "Current transaction rate",
			commonLabels, nil,
		),
		txBlockRate: prometheus.NewDesc(
			"tx_block_rate", "Current TX block rate",
			commonLabels, nil,
		),
		dsBlockRate: prometheus.NewDesc(
			"ds_block_rate", "Current DS block rate",
			commonLabels, nil,
		),
		numPeers: prometheus.NewDesc(
			"num_peers", "Peers count",
			commonLabels, nil,
		),
		numTransactions: prometheus.NewDesc(
			"num_transactions", "Transactions count",
			commonLabels, nil,
		),
		numTxBlocks: prometheus.NewDesc(
			"num_tx_blocks", "Current tx block rate",
			commonLabels, nil,
		),
		numDSBlocks: prometheus.NewDesc(
			"num_ds_blocks", "DS blocks count",
			commonLabels, nil,
		),
		numTxnsTxEpoch: prometheus.NewDesc(
			"num_txns_tx_epoch", "numTxnsTxEpoch",
			commonLabels, nil,
		),
		numTxnsDSEpoch: prometheus.NewDesc(
			"num_txns_ds_epoch", "numTxnsDSEpoch",
			commonLabels, nil,
		),
		shardingPeers: prometheus.NewDesc(
			"sharding_peers", "Peers count of every sharding",
			append([]string{"shard_index"}, commonLabels...), nil,
		),
		difficulty: prometheus.NewDesc(
			"difficulty", "The minimum shard difficulty of the previous block",
			commonLabels, nil,
		),
		dsDifficulty: prometheus.NewDesc(
			"ds_difficulty", "The minimum DS difficulty of the previous block",
			commonLabels, nil,
		),
		networkID: prometheus.NewDesc(
			"network_id", "Network ID of current zilliqa network",
			commonLabels, nil,
		),
		latestTxBlockTimestamp: prometheus.NewDesc(
			"latest_txblock_timestamp", "The timestamp of the latest tx block",
			commonLabels, nil,
		),
		latestDsBlockTimestamp: prometheus.NewDesc(
			"latest_dsblock_timestamp", "The timestamp of the latest ds block",
			commonLabels, nil,
		),
	}
}

func (c *APICollector) Describe(ch chan<- *prometheus.Desc) {
	if !IsGeneralLookup(c.constants.NodeType()) {
		return
	}
	ch <- c.apiServerUp
	ch <- c.epoch
	ch <- c.dsEpoch
	ch <- c.transactionRate
	ch <- c.txBlockRate
	ch <- c.dsBlockRate
	ch <- c.numPeers
	ch <- c.numTransactions
	ch <- c.numTxBlocks
	ch <- c.numDSBlocks
	ch <- c.numTxnsDSEpoch
	ch <- c.numTxnsTxEpoch
	ch <- c.shardingPeers
	ch <- c.difficulty
	ch <- c.dsDifficulty
	ch <- c.networkID
	ch <- c.latestTxBlockTimestamp
}

func (c *APICollector) Collect(ch chan<- prometheus.Metric) {
	labels := c.constants.CommonLabelValues()
	if c.constants.NodeType() == UnknownNodeType {
		log.WithField("endpoint", c.options.APIAddr()).Debug("node type unknown, try to access API Server")
		if err := utils.CheckTCPPortOpen(c.options.APIAddr(), 100*time.Millisecond); err != nil {
			return
		}
		log.WithField("endpoint", c.options.APIAddr()).Info("node type unknown, and API Server not detected")
	} else if !IsGeneralLookup(c.constants.NodeType()) {
		log.Debug("not a lookup server, skip api server info collection")
		return
	}
	log.Debug("enter api collector")
	cli := c.options.GetAPIClient()
	if cli == nil {
		log.Error("API endpoint not set")
		return
	}
	log.Debug("start GetBlockchainInfo")
	info, err := cli.GetBlockchainInfo()
	if err != nil {
		log.WithError(err).Error("error while getting blockchain info")
		ch <- prometheus.MustNewConstMetric(c.apiServerUp, prometheus.GaugeValue, float64(0),
			append([]string{c.options.APIEndpoint()}, labels...)...)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.apiServerUp, prometheus.GaugeValue, float64(1),
		append([]string{c.options.APIEndpoint()}, labels...)...)
	epoch, _ := strconv.ParseFloat(info.CurrentMiniEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.epoch, prometheus.GaugeValue, epoch, labels...)
	dsEpoch, _ := strconv.ParseFloat(info.CurrentDSEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.dsEpoch, prometheus.GaugeValue, dsEpoch, labels...)
	ch <- prometheus.MustNewConstMetric(c.transactionRate, prometheus.GaugeValue, info.TransactionRate, labels...)
	ch <- prometheus.MustNewConstMetric(c.txBlockRate, prometheus.GaugeValue, info.TxBlockRate, labels...)
	ch <- prometheus.MustNewConstMetric(c.dsBlockRate, prometheus.GaugeValue, info.DSBlockRate, labels...)
	ch <- prometheus.MustNewConstMetric(c.numPeers, prometheus.GaugeValue, float64(info.NumPeers), labels...)
	numTransactions, _ := strconv.ParseFloat(info.NumTransactions, 64)
	ch <- prometheus.MustNewConstMetric(c.numTransactions, prometheus.GaugeValue, numTransactions, labels...)
	numTxBlocks, _ := strconv.ParseFloat(info.NumTxBlocks, 64)
	ch <- prometheus.MustNewConstMetric(c.numTxBlocks, prometheus.GaugeValue, numTxBlocks, labels...)
	numDSBlocks, _ := strconv.ParseFloat(info.NumDSBlocks, 64)
	ch <- prometheus.MustNewConstMetric(c.numDSBlocks, prometheus.GaugeValue, numDSBlocks, labels...)
	numTxnsTxEpoch, _ := strconv.ParseFloat(info.NumTxnsTxEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.numTxnsTxEpoch, prometheus.GaugeValue, numTxnsTxEpoch, labels...)
	numTxnsDSEpoch, _ := strconv.ParseFloat(info.NumTxnsDSEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.numTxnsDSEpoch, prometheus.GaugeValue, numTxnsDSEpoch, labels...)
	for i, peers := range info.ShardingStructure.NumPeers {
		ch <- prometheus.MustNewConstMetric(c.shardingPeers, prometheus.GaugeValue, float64(peers),
			append([]string{strconv.Itoa(i)}, labels...)...)
	}
	log.Debug("done GetBlockchainInfo")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		log.Debug("start GetPrevDifficulty")
		defer wg.Done()
		resp, err := cli.GetPrevDifficulty()
		if err != nil {
			log.WithError(err).Error("fail to GetPrevDifficulty")
			return
		}
		ch <- prometheus.MustNewConstMetric(c.difficulty, prometheus.GaugeValue, float64(resp), labels...)
		log.Debug("done GetPrevDifficulty")
	}()

	wg.Add(1)
	go func() {
		log.Debug("start GetPrevDSDifficulty")
		defer wg.Done()
		resp, err := cli.GetPrevDSDifficulty()
		if err != nil {
			log.WithError(err).Error("fail to GetPrevDSDifficulty")
			return
		}
		ch <- prometheus.MustNewConstMetric(c.dsDifficulty, prometheus.GaugeValue, float64(resp), labels...)
		log.Debug("done GetPrevDSDifficulty")
	}()

	wg.Add(1)
	go func() {
		log.Debug("start GetNetworkId")
		defer wg.Done()
		resp, err := cli.GetNetworkId()
		if err != nil {
			log.WithError(err).Error("fail to GetNetworkId")
		}
		netID, err := strconv.ParseFloat(resp, 64)
		if err != nil {
			log.WithError(err).Error("fail to parse GetNetworkId as number")
		}
		ch <- prometheus.MustNewConstMetric(c.networkID, prometheus.GaugeValue, netID, labels...)
		log.Debug("done GetNetworkId")
	}()
	wg.Add(1)
	go func() {
		log.Debug("start GetLatestTxBlock info")
		defer wg.Done()
		block, err := cli.GetLatestTxBlock()
		if err != nil {
			log.WithError(err).Error("fail to GetLatestTxBlock")
		}
		ts, err := strconv.ParseFloat(block.Header.Timestamp, 64)
		if err != nil {
			log.WithError(err).WithField("block", block).Error("fail to parse LatestTxBlock.Header.Timestamp as number")
		}
		ch <- prometheus.MustNewConstMetric(c.latestTxBlockTimestamp, prometheus.GaugeValue, ts/1000, labels...)
		log.Debug("done GetLatestTxBlock")
	}()
	wg.Add(1)
	go func() {
		log.Debug("start GetLatestDsBlock info")
		defer wg.Done()
		block, err := cli.GetLatestDsBlock()
		if err != nil {
			log.WithError(err).Error("fail to GetLatestTxBlock")
		}
		ts, err := strconv.ParseFloat(block.Header.Timestamp, 64)
		if err != nil {
			log.WithError(err).WithField("block", block).Error("fail to parse LatestDsBlock.Header.Timestamp as number")
		}
		ch <- prometheus.MustNewConstMetric(c.latestDsBlockTimestamp, prometheus.GaugeValue, ts/1000, labels...)
		log.Debug("done GetLatestDsBlock")
	}()
	wg.Wait()
	log.Debug("exit api collector")
}
