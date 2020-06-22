package adminclient

import (
	"context"
	"github.com/zilliqa/zilliqa-exporter/jsonrpc"
	"time"
)

type Client struct {
	cli     jsonrpc.Client
	address string
	timeout time.Duration
}

func New(addr string, timeout time.Duration) *Client {
	return &Client{
		cli:     jsonrpc.NewTCPClient(addr),
		timeout: timeout,
	}
}

func (c Client) defaultCtx() (context.Context, context.CancelFunc) {
	if c.timeout == 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), c.timeout)
}

func (c *Client) getResp(ctx context.Context, request *jsonrpc.Request) (*jsonrpc.Response, error) {
	resp, err := c.cli.CallContext(ctx, request)
	if err != nil {
		return nil, err
	}
	err = resp.Err()
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (c *Client) CallBatch(req ...*jsonrpc.Request) ([]*jsonrpc.Response, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.CallBatchContext(ctx, req...)
}

func (c *Client) CallBatchContext(ctx context.Context, req ...*jsonrpc.Request) ([]*jsonrpc.Response, error) {
	return c.cli.CallBatchContext(ctx, req...)
}

func (c *Client) GetCurrentMiniEpoch() (int64, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.GetCurrentMiniEpochContext(ctx)
}

func (c *Client) GetCurrentMiniEpochContext(ctx context.Context) (int64, error) {
	resp, err := c.getResp(ctx, NewGetCurrentMiniEpochReq())
	if err != nil {
		return 0, err
	}
	return resp.GetInt64()
}

func (c *Client) GetCurrentDSEpoch() (int64, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.GetCurrentDSEpochContext(ctx)
}

func (c *Client) GetCurrentDSEpochContext(ctx context.Context) (int64, error) {
	resp, err := c.getResp(ctx, NewGetCurrentDSEpochReq())
	if err != nil {
		return 0, err
	}
	return resp.GetInt64()
}

func (c *Client) GetNodeType() (NodeType, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.GetNodeTypeContext(ctx)
}

func (c *Client) GetNodeTypeContext(ctx context.Context) (NodeType, error) {
	nt := NodeType{}
	resp, err := c.getResp(ctx, NewGetNodeTypeReq())
	if err != nil {
		return NodeType{}, err
	}
	err = resp.GetObject(&nt)
	return nt, err
}

// TODO: GetDSCommittee
//func (c *Client) GetDSCommittee() (int64, error) {
//	ctx, cancel := c.defaultCtx()
//	defer cancel()
//	return c.GetCurrentMiniEpochContext(ctx)
//}
//
//func (c *Client) GetDSCommitteeContext(ctx context.Context) (int64, error) {
//	resp, err := c.getResp(ctx, NewGetDSCommitteeReq())
//	if err != nil {
//		return 0, err
//	}
//	return resp.GetInt64()
//}

func (c *Client) GetNodeState() (NodeState, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.GetNodeStateContext(ctx)
}

func (c *Client) GetNodeStateContext(ctx context.Context) (NodeState, error) {
	var nt NodeState
	resp, err := c.getResp(ctx, NewGetNodeStateReq())
	if err != nil {
		return 0, err
	}
	err = resp.GetObject(&nt)
	return nt, err
}

func (c *Client) GetPrevDifficulty() (int64, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.GetPrevDifficultyContext(ctx)
}

func (c *Client) GetPrevDifficultyContext(ctx context.Context) (int64, error) {
	resp, err := c.getResp(ctx, NewGetPrevDifficultyReq())
	if err != nil {
		return 0, err
	}
	return resp.GetInt64()
}

func (c *Client) GetPrevDSDifficulty() (int64, error) {
	ctx, cancel := c.defaultCtx()
	defer cancel()
	return c.GetPrevDSDifficultyContext(ctx)
}

func (c *Client) GetPrevDSDifficultyContext(ctx context.Context) (int64, error) {
	resp, err := c.getResp(ctx, NewGetPrevDSDifficultyReq())
	if err != nil {
		return 0, err
	}
	return resp.GetInt64()
}

// TODO: the rest methods
