package main

import (
	"flag"
	"fmt"
	"genet_exporter/adminclient"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"strings"
	"time"
)

const (
	Lookup       string = "lookup"
	newLookup    string = "newlookup"
	Level2Lookup string = "level2lookup"

	Normal  string = "normal"
	dsGuard string = "dsguard"
)

type CollectorOptions struct {
	IsMainNet bool
	IsSideCar bool
	IsSameNS  bool

	APIEndpoint          string
	AdminAPIEndpoint     string
	WebsocketAPIEndpoint string

	NodeType string
}

func (c *CollectorOptions) BindFlags(set *flag.FlagSet) {
	set.BoolVar(&c.IsMainNet, "mainnet", false, "run as mainnet mode")
	set.BoolVar(&c.IsSideCar, "sidecar", false, "run as sidecar of zilliqa process")
	set.BoolVar(&c.IsSameNS, "same-ns", false, "run within the same PID namespace of zilliqa process")
	set.StringVar(&c.NodeType, "type", "", "zilliqa node type")
	set.StringVar(&c.APIEndpoint, "api", "127.0.0.1:4201", "zilliqa jsonrpc endpoint")
	set.StringVar(&c.AdminAPIEndpoint, "adminapi", "", "zilliqa admin api endpoint")
	set.StringVar(&c.WebsocketAPIEndpoint, "wsapi", "", "zilliqa websocket api endpoint")
}

func (c CollectorOptions) Endpoint() string {
	if c.APIEndpoint == "" {
		return "http://127.0.0.1:4201"
	}
	if !strings.HasPrefix(c.APIEndpoint, "http://") {
		return fmt.Sprintf("http://%s", c.APIEndpoint)
	}
	return c.APIEndpoint
}

func (c CollectorOptions) AdminEndpoint() string {
	if c.AdminAPIEndpoint != "" {
		return c.AdminAPIEndpoint
	}
	if c.IsSideCar {
		return "127.0.0.1:4301"
	}
	return ""
}

func (c CollectorOptions) CheckAdminEndpoint() error {
	cli := adminclient.New(c.AdminEndpoint(), time.Minute)
	_, err := cli.Call(adminclient.NewRequest("GetCurrentMiniEpoch"))
	return err
}

func (c CollectorOptions) CheckEndpoint() error {
	cli := provider.NewProvider(c.Endpoint())
	_, err := cli.GetCurrentMiniEpoch()
	return err
}

func (c CollectorOptions) GetClient() *provider.Provider {
	return provider.NewProvider(c.Endpoint())
}

func (c CollectorOptions) GetAdminClient() *adminclient.Client {
	return adminclient.New(c.AdminEndpoint(), 5*time.Second)
}
