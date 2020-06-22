package collector

import (
	asserting "github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckEndpoint(t *testing.T) {
	assert := asserting.New(t)
	options := Options{
		IsMainNet:             false,
		NotCollectAPI:         false,
		NotCollectAdmin:       false,
		NotCollectWebsocket:   false,
		NotCollectProcessInfo: false,
		zilliqaBin:            "",
		apiEndpoint:           "127.0.0.1:4201",
		adminEndpoint:         "127.0.0.1:4301",
		websocketEndpoint:     "127.0.0.1:4401",
		nodeType:              "lookup",
	}

	cli, err := options.CheckGetAdminClient()
	t.Log(err)
	assert.NoError(err)
	assert.NotNil(cli)
}
