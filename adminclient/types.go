package adminclient

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
)

type MethodName string

// method names

const (
	GetCurrentMiniEpoch MethodName = "GetCurrentMiniEpoch"
	GetCurrentDSEpoch   MethodName = "GetCurrentDSEpoch"
	GetNodeType         MethodName = "GetNodeType"
	GetDSCommittee      MethodName = "GetDSCommittee"
	GetNodeState        MethodName = "GetNodeState"
	IsTxnInMemPool      MethodName = "IsTxnInMemPool"
	GetPrevDSDifficulty MethodName = "GetPrevDSDifficulty"
	GetPrevDifficulty   MethodName = "GetPrevDifficulty"
	GetSendSCCallsToDS  MethodName = "GetSendSCCallsToDS"

	AddToBlacklistExclusion      MethodName = "AddToBlacklistExclusion"
	RemoveFromBlacklistExclusion MethodName = "RemoveFromBlacklistExclusion"
	ToggleSendSCCallsToDS        MethodName = "ToggleSendSCCallsToDS"
	DisablePoW                   MethodName = "DisablePoW"
	ToggleDisableTxns            MethodName = "ToggleDisableTxns"
)

func ResponseTypeOfMethod(method MethodName) Response {
	switch method {
	case GetCurrentMiniEpoch, GetCurrentDSEpoch, GetPrevDifficulty, GetPrevDSDifficulty:
		return Response(&NumResponse{})
	}
	return Response(&CommonResponse{})
}

func ResponseTypeOfMethods(method ...MethodName) (responses []Response) {
	for _, m := range method {
		responses = append(responses, ResponseTypeOfMethod(m))
	}
	return
}

type Request struct {
	method MethodName
	params []string
}

func NewRequest(method MethodName, params ...string) Request {
	return Request{
		method: method,
		params: params,
	}
}

func ResponseTypeOfRequest(req Request) Response {
	return ResponseTypeOfMethod(req.method)
}

func ResponseTypeOfRequests(req ...Request) (responses []Response) {
	for _, r := range req {
		responses = append(responses, ResponseTypeOfMethod(r.method))
	}
	return
}

type RPCError struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("code: %d %s", e.Code, e.Message)
}

type Response interface {
	Error() error
	Result() interface{}
}

type CommonResponse struct {
	Err *RPCError    `json:"error,omitempty"`
	Rst *interface{} `json:"result,omitempty"`
}

func (r *CommonResponse) Error() error {
	if r.Err == nil {
		return nil
	}
	return r.Err
}

func (r *CommonResponse) Result() interface{} {
	return r.Rst
}

type NumResponse struct {
	Err *RPCError  `json:"error,omitempty"`
	Rst *big.Float `json:"result,omitempty"`
}

func (r *NumResponse) Error() error {
	if r.Err == nil { // golang sucks
		return nil
	}
	return r.Err
}

func (r *NumResponse) Result() interface{} {
	return r.Rst
}

func (r *NumResponse) BigFloat() *big.Float {
	return r.Rst
}

func (r *NumResponse) Float64() float64 {
	f, _ := r.Rst.Float64()
	return f
}

func (r *NumResponse) BigInt() *big.Int {
	i, _ := r.Rst.Int(nil)
	return i
}

func (r *NumResponse) Int64() int64 {
	return r.BigInt().Int64()
}

var UnknownResultType = errors.New("unknown result type")

func (r *NumResponse) UnmarshalJSON(data []byte) error {
	r.Err = nil
	r.Rst = &big.Float{}
	var m map[string]json.RawMessage
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	if e, ok := m["error"]; ok {
		err := json.Unmarshal(e, &r.Err)
		if err != nil {
			return err
		}
	}
	if rst, ok := m["result"]; ok {
		var result interface{}
		err := json.Unmarshal(rst, &result)
		if err != nil {
			return err
		}
		switch v := result.(type) {
		case string:
			_, _, err := r.Rst.Parse(v, 10)
			if err != nil {
				return err
			}
		case float64:
			r.Rst.SetFloat64(v)
		default:
			return UnknownResultType
		}

	}
	return nil
}
