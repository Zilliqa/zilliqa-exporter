package jsonrpc

import (
	"context"
	asserting "github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPClient(t *testing.T) {
	assert := asserting.New(t)
	cli := TCPClient{
		addr:      "dev.zilliqa.com:443",
		closeOnce: false,
		tls:       true,
		tlsConfig: nil,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { cancel() }()
	conn, err := cli.dialContext(ctx)
	assert.Error(err)
	assert.Nil(conn)
	conn, err = cli.dialContext(context.Background())
	assert.NoError(err)
	assert.NotNil(conn)
}
