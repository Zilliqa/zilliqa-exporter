package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
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

	// misc node information, from both admin and lookup api
	// GetNetworkId of api
	networkID *prometheus.Desc
}

func NewAPICollector(constants *Constants) *APICollector {
	constLabels := constants.ConstLabels()
	return &APICollector{
		options:   constants.options,
		constants: constants,
		apiServerUp: prometheus.NewDesc(
			"api_server_up", "JsonRPC server up and running",
			[]string{"endpoint"}, constLabels,
		),
		epoch: prometheus.NewDesc(
			"epoch", "Current TX block number of the node",
			nil, constLabels,
		),
		dsEpoch: prometheus.NewDesc(
			"ds_epoch", "Current DS block number of the node",
			nil, constLabels,
		),
		transactionRate: prometheus.NewDesc(
			"transaction_rate", "Current transaction rate",
			nil, constLabels,
		),
		txBlockRate: prometheus.NewDesc(
			"tx_block_rate", "Current tx block rate",
			nil, constLabels,
		),
		dsBlockRate: prometheus.NewDesc(
			"ds_block_rate", "Current ds block rate",
			nil, constLabels,
		),
		numPeers: prometheus.NewDesc(
			"num_peers", "Peers count",
			nil, constLabels,
		),
		numTransactions: prometheus.NewDesc(
			"num_transactions", "Transactions count",
			nil, constLabels,
		),
		numTxBlocks: prometheus.NewDesc(
			"num_tx_blocks", "Current tx block rate",
			nil, constLabels,
		),
		numDSBlocks: prometheus.NewDesc(
			"num_ds_blocks", "DS blocks count",
			nil, constLabels,
		),
		numTxnsTxEpoch: prometheus.NewDesc(
			"num_txns_tx_epoch", "numTxnsTxEpoch",
			nil, constLabels,
		),
		numTxnsDSEpoch: prometheus.NewDesc(
			"num_txns_ds_epoch", "numTxnsDSEpoch",
			nil, constLabels,
		),
		shardingPeers: prometheus.NewDesc(
			"sharding_peers", "Peers count of every sharding",
			[]string{"index"}, constLabels,
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
	}
}

func (c *APICollector) Describe(ch chan<- *prometheus.Desc) {
	if !c.options.IsGeneralLookup() {
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
}

func (c *APICollector) Collect(ch chan<- prometheus.Metric) {
	if !c.options.IsGeneralLookup() {
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
		ch <- prometheus.MustNewConstMetric(c.apiServerUp, prometheus.GaugeValue, float64(0), c.options.APIEndpoint())
		return
	}
	ch <- prometheus.MustNewConstMetric(c.apiServerUp, prometheus.GaugeValue, float64(1), c.options.APIEndpoint())
	epoch, _ := strconv.ParseFloat(info.CurrentMiniEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.epoch, prometheus.GaugeValue, epoch)
	dsEpoch, _ := strconv.ParseFloat(info.CurrentDSEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.dsEpoch, prometheus.GaugeValue, dsEpoch)
	ch <- prometheus.MustNewConstMetric(c.transactionRate, prometheus.GaugeValue, info.TransactionRate)
	ch <- prometheus.MustNewConstMetric(c.txBlockRate, prometheus.GaugeValue, info.TxBlockRate)
	ch <- prometheus.MustNewConstMetric(c.dsBlockRate, prometheus.GaugeValue, info.DSBlockRate)
	ch <- prometheus.MustNewConstMetric(c.numPeers, prometheus.GaugeValue, float64(info.NumPeers))
	numTransactions, _ := strconv.ParseFloat(info.NumTransactions, 64)
	ch <- prometheus.MustNewConstMetric(c.numTransactions, prometheus.GaugeValue, numTransactions)
	numTxBlocks, _ := strconv.ParseFloat(info.NumTxBlocks, 64)
	ch <- prometheus.MustNewConstMetric(c.numTxBlocks, prometheus.GaugeValue, numTxBlocks)
	numDSBlocks, _ := strconv.ParseFloat(info.NumDSBlocks, 64)
	ch <- prometheus.MustNewConstMetric(c.numDSBlocks, prometheus.GaugeValue, numDSBlocks)
	numTxnsTxEpoch, _ := strconv.ParseFloat(info.NumTxnsTxEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.numTxnsTxEpoch, prometheus.GaugeValue, numTxnsTxEpoch)
	numTxnsDSEpoch, _ := strconv.ParseFloat(info.NumTxnsDSEpoch, 64)
	ch <- prometheus.MustNewConstMetric(c.numTxnsDSEpoch, prometheus.GaugeValue, numTxnsDSEpoch)
	for i, peers := range info.ShardingStructure.NumPeers {
		ch <- prometheus.MustNewConstMetric(c.shardingPeers, prometheus.GaugeValue, float64(peers), strconv.Itoa(i))
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
		ch <- prometheus.MustNewConstMetric(c.difficulty, prometheus.GaugeValue, float64(resp))
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
		ch <- prometheus.MustNewConstMetric(c.dsDifficulty, prometheus.GaugeValue, float64(resp))
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
		ch <- prometheus.MustNewConstMetric(c.networkID, prometheus.GaugeValue, netID)
		log.Debug("done GetNetworkId")
	}()
	wg.Wait()
	log.Debug("exit api collector")
}
