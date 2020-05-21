package main

import "math/big"

type BlockchainInfo struct {
	CurrentDSEpoch    int64
	CurrentMiniEpoch  int64

	// TODO: following
	DSBlockRate       big.Float
	NumDSBlocks       int64
	NumPeers          int
	NumTransactions   int64
	NumTxBlocks       int64
	NumTxnsDSEpoch    int64
	NumTxnsTxEpoch    int64
	ShardingStructure ShardingStructure
	TransactionRate   big.Float
	TxBlockRate       big.Float
}

type ShardingStructure struct {
	NumPeers []int
}
