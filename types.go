package main

import "math/big"

type BlockchainInfo struct {
	CurrentDSEpoch    string
	CurrentMiniEpoch  string

	// TODO: following
	DSBlockRate       big.Float
	NumDSBlocks       string
	NumPeers          int
	NumTransactions   string
	NumTxBlocks       string
	NumTxnsDSEpoch    string
	NumTxnsTxEpoch    string
	ShardingStructure ShardingStructure
	TransactionRate   big.Float
	TxBlockRate       big.Float
}

type ShardingStructure struct {
	NumPeers []int
}
