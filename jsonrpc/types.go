package jsonrpc

import (
	"encoding/json"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

type Request struct {
	id     int64
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func NewRequest(method string, params interface{}) *Request {
	return &Request{Method: method, Params: params}
}

func (r Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      r.id,
		"method":  r.Method,
		"params":  r.Params,
	})
}

func parseResponse(raw []byte) (*Response, error) {
	resp := &Response{}
	err := json.Unmarshal(raw, resp)
	if err != nil {
		err = resp.check()
	}
	return resp, err
}

func parseBatchResponse(raw []byte) ([]*Response, error) {
	var resps []*Response
	err := json.Unmarshal(raw, &resps)
	if err != nil {
		var errs []error
		for _, resp := range resps {
			if e := resp.check(); e != nil {
				errs = append(errs, e)
			}
		}
		var errStrings []string
		for _, e := range errs {
			errStrings = append(errStrings, e.Error())
		}
		if len(errs) == 1 {
			err = errors.Wrap(errs[0], "jsonrpc response parse error")
		} else if len(errs) >= 1 {
			err = errors.New("jsonrpc response parse error:" + strings.Join(errStrings, ";"))
		}
	}
	return resps, err
}

type Response struct {
	Version string          `json:"jsonrpc,omitempty"`
	Result  json.RawMessage `json:"error,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	Id      int64           `json:"id,omitempty"`
}

// TODO: response validation
func (r Response) check() error {
	return nil
}

// used for error assertion
func (r *Response) Err() error {
	if r.Error == nil {
		return nil
	}
	return r.Error
}

func (r *Response) RawResult() []byte {
	return r.Result
}

func (r *Response) GetObject(obj interface{}) error {
	return json.Unmarshal(r.Result, obj)
}

func (r *Response) GetString() (string, error) {
	var v string
	err := r.GetObject(&v)
	return v, err
}

func (r *Response) GetInt() (int, error) {
	var v int
	err := r.GetObject(&v)
	return v, err
}

func (r *Response) GetInt64() (int64, error) {
	var v int64
	err := r.GetObject(&v)
	return v, err
}

func (r *Response) GetFloat64() (float64, error) {
	var v float64
	err := r.GetObject(&v)
	return v, err
}

func (r *Response) GetBigFloat() (big.Float, error) {
	var v big.Float
	err := r.GetObject(&v)
	return v, err
}
