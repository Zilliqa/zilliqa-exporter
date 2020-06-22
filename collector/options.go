package collector

import (
	"encoding/json"
	"fmt"
	"genet_exporter/adminclient"
	"genet_exporter/utils"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

var DetectedNodeType NodeType

type NodeType int

const (
	Lookup NodeType = iota
	newLookup
	Level2Lookup
	Normal
	DSGuard
	UnknownNodeType
)

var nodeTypeStringMap = map[NodeType]string{
	Lookup:          "lookup",
	newLookup:       "newlookup",
	Level2Lookup:    "level2lookup",
	Normal:          "normal",
	DSGuard:         "dsguard",
	UnknownNodeType: "",
}

var (
	lookUpTypes = []NodeType{Lookup, newLookup, Level2Lookup}
	nodeTypes   = []NodeType{Lookup, newLookup, Level2Lookup, Normal, DSGuard}
)

func (n NodeType) String() string {
	if s, ok := nodeTypeStringMap[n]; ok {
		return s
	}
	return ""
}

func NodeTypeFromString(typ string) NodeType {
	for _, t := range nodeTypes {
		if strings.EqualFold(t.String(), typ) {
			return t
		}
	}
	return UnknownNodeType
}

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

	apiEndpoint       string
	adminEndpoint     string
	websocketEndpoint string

	rpcTimeout time.Duration

	// TODO: detect node type from process cmdline
	nodeType string
}

func (c *Options) BindFlags(set *pflag.FlagSet) {
	set.SortFlags = false
	set.BoolVar(&c.IsMainNet, "mainnet", false, "collect mainnet metrics")
	set.BoolVarP(&c.NotCollectAPI, "not-collect-api", "a", false, "do not collect metrics from JSONRPC API")
	set.BoolVarP(&c.NotCollectAdmin, "not-collect-admin", "m", false, "do not collect metrics from Admin API")
	set.BoolVarP(&c.NotCollectWebsocket, "not-collect-websocket", "w", false, "do not collect metrics from Websocket API")
	set.BoolVarP(&c.NotCollectProcessInfo, "not-collect-process-info", "p", false, "do not collect metrics from Zilliqa Process")
	set.DurationVarP(&c.rpcTimeout, "rpc-timeout", "t", 30*time.Second, "timeout of rpc request")
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
	if c.apiEndpoint == "" && utils.CheckTCPPortOpen(DefaultAPIEndpoint) == nil {
		c.apiEndpoint = DefaultAPIEndpoint
	}
	if !strings.HasPrefix(c.apiEndpoint, "http://") && !strings.HasPrefix(c.apiEndpoint, "https://") {
		return fmt.Sprintf("http://%s", c.apiEndpoint)
	}
	return c.apiEndpoint
}

func (c Options) AdminEndpoint() string {
	if c.adminEndpoint == "" && utils.CheckTCPPortOpen(DefaultAdminEndpoint) == nil {
		c.adminEndpoint = DefaultAdminEndpoint
	}
	return c.adminEndpoint
}

func (c Options) WebsocketEndpoint() string {
	if c.websocketEndpoint == "" && utils.CheckTCPPortOpen(DefaultWebsocketEndpoint) == nil {
		c.websocketEndpoint = DefaultWebsocketEndpoint
	}
	return c.websocketEndpoint
}

func (c Options) CheckGetAPIClient() (*provider.Provider, error) {
	ep := c.APIEndpoint()
	if ep == "" {
		return nil, errors.New("api endpoint not set")
	}
	u, err := url.Parse(ep)
	if err != nil {
		return nil, err
	}
	host := u.Host
	if len(strings.Split(host, ":")) != 2 {
		if u.Scheme == "http" {
			host = host + ":80"
		} else if u.Scheme == "https" {
			host = host + ":443"
		}
	}
	if err := utils.CheckTCPPortOpen(host); err != nil {
		return nil, errors.Wrap(err, "cannot connect to api server")
	}
	cli := c.GetAPIClient()
	_, err = cli.GetCurrentMiniEpoch()
	return cli, err
}

func (c Options) CheckGetAdminClient() (*adminclient.Client, error) {
	ep := c.AdminEndpoint()
	if ep == "" {
		return nil, errors.New("admin endpoint not set")
	}
	if err := utils.CheckTCPPortOpen(ep); err != nil {
		return nil, errors.Wrap(err, "cannot connect to admin server")
	}
	cli := c.GetAdminClient()
	_, err := cli.GetCurrentMiniEpoch()
	return cli, err
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

func (c Options) IsGeneralLookup() bool {
	var isLookup bool
	for _, typ := range lookUpTypes {
		if strings.EqualFold(string(c.NodeType()), string(typ)) {
			isLookup = true
		}
	}
	return isLookup
}

func (c Options) NodeType() NodeType {
	if c.nodeType == "" {
		return DetectedNodeType
	}
	return NodeTypeFromString(c.nodeType)
}

func (c *Options) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"IsMainnet":             c.IsMainNet,
		"NotCollectAPI":         c.NotCollectAPI,
		"NotCollectAdmin":       c.NotCollectAdmin,
		"NotCollectWebsocket":   c.NotCollectWebsocket,
		"NotCollectProcessInfo": c.NotCollectProcessInfo,
		"ZilliqaBinPath":        c.ZilliqaBinPath(),
		"ApiEndpoint":           c.APIEndpoint(),
		"AdminEndpoint":         c.AdminEndpoint(),
		"WebsocketEndpoint":     c.WebsocketEndpoint(),
		"RpcTimeout":            c.rpcTimeout.String(),
		"NodeType":              c.NodeType().String(),
	})
}
