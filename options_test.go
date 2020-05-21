package main

import (
	asserting "github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckEndpoint(t *testing.T) {
	assert := asserting.New(t)
	options := CollectorOptions{
		IsMainNet:            false,
		IsSideCar:            false,
		IsSameNS:             false,
		APIEndpoint:          "127.0.0.1:4201",
		AdminAPIEndpoint:     "127.0.0.1:4301",
		WebsocketAPIEndpoint: "",
		NodeType:             "",
	}

	err := options.CheckAdminEndpoint()
	t.Log(err)
	assert.NoError(err)
}
