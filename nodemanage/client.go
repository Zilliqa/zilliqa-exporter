package nodemanage

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Client struct {
	address string
	timeout time.Duration
}

func NewClient(addr string, timeout time.Duration) *Client {
	return &Client{address: addr, timeout: timeout}
}

func (m *Client) dial() (net.Conn, error) {
	return net.DialTimeout("tcp", m.address, m.timeout)
}

func (m *Client) getRawResp(data []byte) ([]byte, error) {
	conn, err := m.dial()
	if err != nil {
		return nil, errors.Wrap(err, "fail to connect to server")
	}
	defer conn.Close()
	log.WithField("addr", m.address).WithField("payload", string(data)).Debug("sending request")
	_, err = conn.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to send data to server")
	}
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, conn)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get response")
	}
	return buf.Bytes(), nil
}

func (m *Client) getPayload(rq ...Request) []byte {
	if len(rq) == 0 {
		panic("empty requests")
	}
	requests := make([]map[string]interface{}, 16)
	for _, r := range rq {
		params := r.params
		if r.params == nil || len(r.params) == 0 {
			params = []string{""}
		}
		requests = append(requests, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  r.method,
			"params":  params,
		})
	}
	var pl []byte
	if len(rq) == 1 {
		pl, _ = json.Marshal(requests[0])
	} else {
		pl, _ = json.Marshal(requests)
	}
	return append(pl, []byte("\n")...)
}

func parseResponse(data []byte, resp Response) error {
	return json.Unmarshal(data, resp)
}

func parseBatchResponse(data []byte, responses []Response) error {
	var rawResps []json.RawMessage
	err := json.Unmarshal(data, &rawResps)
	if err != nil {
		return err
	}
	for i, r := range rawResps {
		err := parseResponse(r, responses[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Client) Call(request Request) (Response, error) {
	payload := m.getPayload(request)
	rawResp, err := m.getRawResp(payload)
	if err != nil {
		return nil, err
	}
	resp := ResponseTypeOfRequest(request)
	err = parseResponse(rawResp, resp)
	return resp, err
}

func (m *Client) CallBatch(requests ... Request) ([]Response, error) {
	payload := m.getPayload(requests...)
	rawResp, err := m.getRawResp(payload)
	if err != nil {
		return nil, err
	}
	resps := ResponseTypeOfRequests(requests...)
	err = parseBatchResponse(rawResp, resps)
	return resps, err
}
