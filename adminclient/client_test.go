package adminclient

import (
	"net"
	"reflect"
	"testing"
	"time"
)
import asserting "github.com/stretchr/testify/assert"

func testServer(output []byte) (net.Addr, error) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		return nil, err
	}
	go func() {
		defer l.Close()
		conn, err := l.AcceptTCP()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Read(make([]byte, 4096))
		_, _ = conn.Write(output)
	}()
	return l.Addr(), nil
}

func TestClient(t *testing.T) {
	assert := asserting.New(t)
	resp := ResponseTypeOfMethod(GetCurrentMiniEpoch)
	assert.NoError(parseResponse([]byte(`{"id":"1","jsonrpc":"2.0","result":"1934"}`), resp))
	assert.IsType(&NumResponse{}, resp)
	assert.Equal(resp.(*NumResponse).Int64(), int64(1934))

	resps := ResponseTypeOfMethods(GetCurrentMiniEpoch, GetCurrentDSEpoch)
	assert.NoError(parseBatchResponse(
		[]byte(`[{"id":"1","jsonrpc":"2.0","result":"1934"},{"id":"1","jsonrpc":"2.0","result":"1935"}]`),
		resps,
	))
	assert.Equal(resps[0].(*NumResponse).Int64(), int64(1934))
	assert.Equal(resps[1].(*NumResponse).Int64(), int64(1935))

	addr, err := testServer([]byte(`{"id":"1","jsonrpc":"2.0","result":"1934"}`))
	assert.NoError(err)
	cli := New(addr.String(), 10*time.Second)
	resp, err = cli.Call(NewRequest(GetCurrentDSEpoch))
	assert.NoError(err)
	t.Log(nil == error(nil), error(nil) == resp.Error(), reflect.TypeOf(resp.Error()))
	assert.NoError(resp.Error())
	assert.IsType(&NumResponse{}, resp)
	assert.Equal(resp.(*NumResponse).Int64(), int64(1934))

	addr, err = testServer([]byte(`[{"id":"1","jsonrpc":"2.0","result":"1934"},{"id":"1","jsonrpc":"2.0","result":"1935"}]`))
	assert.NoError(err)
	cli = New(addr.String(), 10*time.Second)
	resps, err = cli.CallBatch(NewRequest(GetCurrentDSEpoch), NewRequest(GetCurrentMiniEpoch))
	assert.NoError(err)
	t.Log(nil == error(nil), error(nil) == resp.Error(), reflect.TypeOf(resp.Error()))
	assert.NoError(resp.Error())
	assert.IsType(&NumResponse{}, resps[0])
	assert.Equal(resps[0].(*NumResponse).Int64(), int64(1934))
	assert.Equal(resps[1].(*NumResponse).Int64(), int64(1935))
}
