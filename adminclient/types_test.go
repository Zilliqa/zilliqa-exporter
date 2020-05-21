package adminclient

import (
	"encoding/json"
	asserting "github.com/stretchr/testify/assert"
	"testing"
)

func TestNumType(t *testing.T) {
	assert := asserting.New(t)
	var resp = NumResponse{}
	err := json.Unmarshal(
		[]byte(`{"error":{"code":-32600,"data":null,"message":"INVALID_JSON_REQUEST: The JSON sent is not a valid JSON-RPC Request object: Not to be queried on lookup"},"id":"1","jsonrpc":"2.0"}`),
		&resp,
	)
	assert.NoError(err)
	assert.Error(resp.Err)
	assert.Error(resp.Error())
	assert.Equal(resp.Err.Code, int64(-32600))
	assert.Equal(resp.Err.Message, "INVALID_JSON_REQUEST: The JSON sent is not a valid JSON-RPC Request object: Not to be queried on lookup")
	assert.Nil(resp.Err.Data)

	err = json.Unmarshal([]byte(`{"id":"1","jsonrpc":"2.0","result":"1934"}`), &resp)
	assert.NoError(err)
	assert.Nil(resp.Error())
	assert.Equal(resp.Int64(), int64(1934))

	err = json.Unmarshal([]byte(`{"id":"1","jsonrpc":"2.0","result":1934}`), &resp)
	assert.Equal(resp.Int64(), int64(1934))

	r := ResponseTypeOfMethod(GetCurrentMiniEpoch)
	assert.IsType(&NumResponse{}, r)
}
