package adminclient

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zilliqa/zilliqa-exporter/jsonrpc"
	"strconv"
	"strings"
)

type MethodName string

// method names
var UnknownResultType = errors.New("unknown result type")

const (
	GetCurrentMiniEpoch MethodName = "GetCurrentMiniEpoch"
	GetCurrentDSEpoch   MethodName = "GetCurrentDSEpoch"
	GetNodeType         MethodName = "GetNodeType"
	GetDSCommittee      MethodName = "GetDSCommittee"
	GetNodeState        MethodName = "GetNodeState"
	GetPrevDifficulty   MethodName = "GetPrevDifficulty"
	GetPrevDSDifficulty MethodName = "GetPrevDSDifficulty"
	GetSendSCCallsToDS  MethodName = "GetSendSCCallsToDS"

	IsTxnInMemPool MethodName = "IsTxnInMemPool"

	AddToBlacklistExclusion      MethodName = "AddToBlacklistExclusion"
	RemoveFromBlacklistExclusion MethodName = "RemoveFromBlacklistExclusion"
	ToggleSendSCCallsToDS        MethodName = "ToggleSendSCCallsToDS"
	DisablePoW                   MethodName = "DisablePoW"
	ToggleDisableTxns            MethodName = "ToggleDisableTxns"
)

func NewReq(method MethodName, params interface{}) *jsonrpc.Request {
	return jsonrpc.NewRequest(string(method), params)
}

func NewGetCurrentMiniEpochReq() *jsonrpc.Request {
	return NewReq(GetCurrentMiniEpoch, nil)
}

func NewGetCurrentDSEpochReq() *jsonrpc.Request {
	return NewReq(GetCurrentDSEpoch, nil)
}

func NewGetNodeTypeReq() *jsonrpc.Request {
	return NewReq(GetNodeType, nil)
}

func NewGetDSCommitteeReq() *jsonrpc.Request {
	return NewReq(GetDSCommittee, nil)
}

func NewGetNodeStateReq() *jsonrpc.Request {
	return NewReq(GetNodeState, nil)
}

func NewIsTxnInMemPoolReq(txn string) *jsonrpc.Request {
	return NewReq(IsTxnInMemPool, []string{txn})
}

func NewGetPrevDifficultyReq() *jsonrpc.Request {
	return NewReq(GetPrevDifficulty, nil)
}

func NewGetSendSCCallsToDSReq() *jsonrpc.Request {
	return NewReq(GetSendSCCallsToDS, nil)
}

func NewGetPrevDSDifficultyReq() *jsonrpc.Request {
	return NewReq(GetPrevDSDifficulty, nil)
}

func NewAddToBlacklistExclusionReq(ip string) *jsonrpc.Request {
	return NewReq(AddToBlacklistExclusion, []string{ip})
}

func NewRemoveFromBlacklistExclusionReq(ip string) *jsonrpc.Request {
	return NewReq(RemoveFromBlacklistExclusion, []string{ip})
}

func NewToggleSendSCCallsToDSReq() *jsonrpc.Request {
	return NewReq(ToggleSendSCCallsToDS, nil)
}

func NewDisablePoWReq() *jsonrpc.Request {
	return NewReq(DisablePoW, nil)
}

func NewToggleDisableTxnsReq() *jsonrpc.Request {
	return NewReq(ToggleDisableTxns, nil)
}

type NodeTypeName int

const (
	NotInNetwork NodeTypeName = iota
	Seed
	Lookup
	DSNode
	ShardNode
)

func (n NodeTypeName) String() string {
	switch n {
	case NotInNetwork:
		return "NotInNetwork"
	case Seed:
		return "Seed"
	case Lookup:
		return "Lookup"
	case DSNode:
		return "DSNode"
	case ShardNode:
		return "ShardNode"
	}
	return "Unknown NodeType"
}

type NodeType struct {
	Type      NodeTypeName
	ShardId   int
	TillEpoch int
}

func (n NodeType) String() string {
	switch n.Type {
	case NotInNetwork:
		return fmt.Sprintf("Not in network, synced till epoch %d", n.TillEpoch)
	case Seed:
		return "Seed"
	case Lookup:
		return "Lookup"
	case DSNode:
		return "DS Node"
	case ShardNode:
		return fmt.Sprintf("Shard Node of shard %d", n.ShardId)
	}
	return "Unknown NodeType"
}

func (n *NodeType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch {
	case strings.HasPrefix(s, "Not in network"):
		split := strings.Split(s, " ")
		epoch, err := strconv.Atoi(split[len(split)-1])
		if err != nil {
			return err
		}
		n.Type = NotInNetwork
		n.TillEpoch = epoch
	case strings.HasPrefix(s, "Shard Node of"):
		split := strings.Split(s, " ")
		num, err := strconv.Atoi(split[len(split)-1])
		if err != nil {
			return err
		}
		n.Type = ShardNode
		n.ShardId = num
	case s == "DS Node":
		n.Type = DSNode
	case s == "Lookup":
		n.Type = Lookup
	case s == "Seed":
		n.Type = Seed
	default:
		return errors.New("parse NodeTypeName error: unknown node type " + s)
	}
	return nil
}

type NodeState int

const (
	POW_SUBMISSION NodeState = iota
	DSBLOCK_CONSENSUS_PREP
	DSBLOCK_CONSENSUS
	MICROBLOCK_SUBMISSION
	FINALBLOCK_CONSENSUS_PREP
	FINALBLOCK_CONSENSUS
	VIEWCHANGE_CONSENSUS_PREP
	VIEWCHANGE_CONSENSUS
	ERROR
	SYNC
)

var StringNodeStateMap = map[string]NodeState{
	"POW_SUBMISSION":            POW_SUBMISSION,
	"DSBLOCK_CONSENSUS_PREP":    DSBLOCK_CONSENSUS_PREP,
	"DSBLOCK_CONSENSUS":         DSBLOCK_CONSENSUS,
	"MICROBLOCK_SUBMISSION":     MICROBLOCK_SUBMISSION,
	"FINALBLOCK_CONSENSUS_PREP": FINALBLOCK_CONSENSUS_PREP,
	"FINALBLOCK_CONSENSUS":      FINALBLOCK_CONSENSUS,
	"VIEWCHANGE_CONSENSUS_PREP": VIEWCHANGE_CONSENSUS_PREP,
	"VIEWCHANGE_CONSENSUS":      VIEWCHANGE_CONSENSUS,
	"ERROR":                     ERROR,
	"SYNC":                      SYNC,
}

var NodeStateStringMap = func(m map[string]NodeState) map[NodeState]string {
	rt := make(map[NodeState]string)
	for k, v := range m {
		rt[v] = k
	}
	return rt
}(StringNodeStateMap)

func (n NodeState) String() string {
	return NodeStateStringMap[n]
}

func (n *NodeState) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if state, ok := StringNodeStateMap[s]; ok {
		*n = state
		return nil
	}
	return errors.New("parse NodeState error: unknown state " + s)
}
