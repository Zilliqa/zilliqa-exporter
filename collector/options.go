package collector

import (
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/zilliqa/zilliqa-exporter/adminclient"
	"github.com/zilliqa/zilliqa-exporter/utils"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultAPIEndpoint       = "127.0.0.1:4201"
	DefaultAdminEndpoint     = "127.0.0.1:4301"
	DefaultWebsocketEndpoint = "127.0.0.1:4401"
)

type Options struct {
	IsMainNet bool

	NotCollectAPI         bool
	NotCollectAdmin       bool
	NotCollectWebsocket   bool
	NotCollectProcessInfo bool

	zilliqaBin string

	p2pPort           uint32
	apiEndpoint       string
	adminEndpoint     string
	websocketEndpoint string

	rpcTimeout time.Duration

	nodeType string
}

func (c *Options) BindFlags(set *pflag.FlagSet) {
	set.SortFlags = false
	set.BoolVar(&c.IsMainNet, "mainnet", false, "collect mainnet metrics")
	set.BoolVar(&c.NotCollectAPI, "not-collect-api", false, "do not collect metrics from JSONRPC API")
	set.BoolVar(&c.NotCollectAdmin, "not-collect-admin", false, "do not collect metrics from Admin API")
	set.BoolVar(&c.NotCollectWebsocket, "not-collect-websocket", false, "do not collect metrics from Websocket API")
	set.BoolVar(&c.NotCollectProcessInfo, "not-collect-process-info", false, "do not collect metrics from Zilliqa Process")
	set.DurationVarP(&c.rpcTimeout, "rpc-timeout", "t", 10*time.Second, "timeout of rpc request")
	set.Uint32Var(&c.p2pPort, "p2p-port", 33133, "p2p port of zilliqa node")
	set.StringVar(&c.apiEndpoint, "api", "", "zilliqa jsonrpc endpoint")
	set.StringVar(&c.adminEndpoint, "admin", "", "zilliqa admin api endpoint")
	set.StringVar(&c.websocketEndpoint, "ws", "", "zilliqa websocket api endpoint")
	set.StringVar(&c.zilliqaBin, "bin", "zilliqa", "the zilliqa executable name or path")
	set.StringVar(&c.nodeType, "type", "", "zilliqa node type")
}

func (c *Options) ZilliqaBinPath() string {
	if c.zilliqaBin == "" {
		return ""
	}
	if utils.PathIsFile(c.zilliqaBin) {
		return c.zilliqaBin
	}
	path, err := exec.LookPath(c.zilliqaBin)
	if err != nil {
		log.WithField("zilliqaBin", c.zilliqaBin).WithError(err).Error("zilliqa executable not found")
		return ""
	}
	c.zilliqaBin = path
	return path
}

func (c Options) APIEndpoint() string {
	//if c.apiEndpoint == "" && utils.CheckTCPPortOpen(DefaultAPIEndpoint) == nil {
	if c.apiEndpoint == "" {
		c.apiEndpoint = DefaultAPIEndpoint
	}
	if !strings.HasPrefix(c.apiEndpoint, "http://") && !strings.HasPrefix(c.apiEndpoint, "https://") {
		return fmt.Sprintf("http://%s", c.apiEndpoint)
	}
	return c.apiEndpoint
}

func (c Options) AdminEndpoint() string {
	//if c.adminEndpoint == "" && utils.CheckTCPPortOpen(DefaultAdminEndpoint) == nil {
	if c.adminEndpoint == "" {
		c.adminEndpoint = DefaultAdminEndpoint
	}
	return c.adminEndpoint
}

func (c Options) WebsocketEndpoint() string {
	//if c.websocketEndpoint == "" && utils.CheckTCPPortOpen(DefaultWebsocketEndpoint) == nil {
	if c.websocketEndpoint == "" {
		c.websocketEndpoint = DefaultWebsocketEndpoint
	}
	return c.websocketEndpoint
}

func (c Options) GetAPIClient() *provider.Provider {
	ep := c.APIEndpoint()
	if ep == "" {
		return nil
	}
	return provider.NewProvider(ep)
}

func (c Options) GetAdminClient() *adminclient.Client {
	ep := c.AdminEndpoint()
	if ep == "" {
		return nil
	}
	return adminclient.New(ep, c.rpcTimeout)
}

func (c *Options) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.ToMap())
}

func (c *Options) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"IsMainnet":             c.IsMainNet,
		"NotCollectAPI":         c.NotCollectAPI,
		"NotCollectAdmin":       c.NotCollectAdmin,
		"NotCollectWebsocket":   c.NotCollectWebsocket,
		"NotCollectProcessInfo": c.NotCollectProcessInfo,
		"ZilliqaBinPath":        c.ZilliqaBinPath(),
		"p2pPort":               c.p2pPort,
		"ApiEndpoint":           c.APIEndpoint(),
		"AdminEndpoint":         c.AdminEndpoint(),
		"WebsocketEndpoint":     c.WebsocketEndpoint(),
		"RpcTimeout":            c.rpcTimeout.String(),
		"NodeType":              c.nodeType,
	}
}
