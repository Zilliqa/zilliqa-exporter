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
		apiEndpoint:           "https://api.zilliqa.com",
		adminEndpoint:         "127.0.0.1:4301",
		websocketEndpoint:     "127.0.0.1:4401",
		nodeType:              "lookup",
	}

	cli := options.GetAPIClient()
	assert.NotNil(cli)
}
