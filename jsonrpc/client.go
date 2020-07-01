package jsonrpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync/atomic"
)

type Conn net.Conn

type Client interface {
	Address() string
	Call(request *Request) (*Response, error)
	CallContext(ctx context.Context, request *Request) (*Response, error)
	CallBatch(requests ...*Request) ([]*Response, error)
	CallBatchContext(ctx context.Context, requests ...*Request) ([]*Response, error)
}

type TCPClient struct {
	addr      string
	closeOnce bool
	counter   int64
	// TLS
	tls       bool
	tlsConfig *tls.Config
}

func NewTCPClient(addr string) *TCPClient {
	return &TCPClient{addr: addr}
}

func (c TCPClient) Address() string {
	return c.addr
}

func (c *TCPClient) getId() int64 {
	return atomic.AddInt64(&c.counter, 1)
}

func (c *TCPClient) setReqId(rq ...*Request) {
	for _, r := range rq {
		r.id = c.getId()
	}
}

func (c *TCPClient) dial() (net.Conn, error) {
	return c.dialContext(context.Background())
}

func (c *TCPClient) dialContext(ctx context.Context) (net.Conn, error) {
	var d net.Dialer
	if c.tls {
		ch := make(chan net.Conn)
		errCh := make(chan error)
		go func() {
			conn, err := tls.Dial("tcp", c.addr, c.tlsConfig)
			if err != nil {
				errCh <- err
				if conn != nil {
					_ = conn.Close()
				}
				return
			}
			ch <- conn
		}()

		select {
		case conn := <-ch:
			return conn, nil
		case err := <-errCh:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return d.DialContext(ctx, "tcp", c.addr)
}

func (c *TCPClient) getRawResp(ctx context.Context, data []byte) ([]byte, error) {
	conn, err := c.dialContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to connect to server")
	}
	defer conn.Close()
	log.WithField("addr", c.addr).WithField("payload", string(data)).Debug("sending request")
	_, err = conn.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to send data to server")
	}
	buf := new(bytes.Buffer)

	doneCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		_, err = io.Copy(buf, conn)
		log.Debug("reading response")
		if err != nil {
			errCh <- errors.Wrap(err, "fail to get response")
		}
		close(doneCh)
		close(errCh)
	}()
	select {
	case err := <-errCh:
		log.WithField("addr", c.addr).WithError(err).Error("connection Error")
		return nil, err
	case <-ctx.Done():
		log.WithField("addr", c.addr).Error("connection timeout or canceled")
		return nil, ctx.Err()
	case <-doneCh:
	}
	<-doneCh
	res := buf.Bytes()
	log.WithField("addr", c.addr).WithField("resp", string(res)).Debug("got response")
	return buf.Bytes(), nil
}

func (*TCPClient) getBatchPayload(rq ...*Request) ([]byte, error) {
	if len(rq) == 0 {
		return nil, errors.New("empty requests")
	}
	d, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}
	d = append(d, byte('\n'))
	return d, nil
}

func (*TCPClient) getPayload(rq *Request) ([]byte, error) {
	d, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}
	d = append(d, byte('\n'))
	return d, nil
}

func (c *TCPClient) Call(request *Request) (*Response, error) {
	return c.CallContext(context.Background(), request)
}

func (c *TCPClient) CallContext(ctx context.Context, request *Request) (*Response, error) {
	payload, err := c.getPayload(request)
	if err != nil {
		return nil, err
	}
	rawResp, err := c.getRawResp(ctx, payload)
	if err != nil {
		return nil, err
	}
	return parseResponse(rawResp)
}

func (c *TCPClient) CallBatch(requests ...*Request) ([]*Response, error) {
	return c.CallBatchContext(context.Background(), requests...)
}

func (c *TCPClient) CallBatchContext(ctx context.Context, requests ...*Request) ([]*Response, error) {
	payload, err := c.getBatchPayload(requests...)
	if err != nil {
		return nil, err
	}
	rawResp, err := c.getRawResp(ctx, payload)
	if err != nil {
		return nil, err
	}
	return parseBatchResponse(rawResp)
}
