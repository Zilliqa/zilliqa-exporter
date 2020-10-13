package collector

import "strings"

type NodeType int

const (
	UnknownNodeType NodeType = iota
	Lookup
	Seed    // dedicated seeds
	SeedPub // guard seeds that exposes api to public
	SeedPrv // guard seeds that exposes api to some specific client
	Normal
	DSGuard
)

const (
	Level2Lookup = SeedPub
	NewLookup    = SeedPrv
)

var stringNodeTypeMap = map[string]NodeType{
	"":             UnknownNodeType,
	"lookup":       Lookup,
	"seed-apipub":  SeedPub,
	"seed":         Seed,
	"apipub":       SeedPub,
	"seedpub":      SeedPub,
	"level2lookup": SeedPub,
	"seed-apiprv":  SeedPrv,
	"apiprv":       SeedPrv,
	"seedprv":      SeedPrv,
	"newlookup":    SeedPrv,
	"normal":       Normal,
	"dsguard":      DSGuard,
}

var nodeTypeStringMap = map[NodeType]string{
	UnknownNodeType: "",
	Lookup:          "lookup",
	Seed:            "seed",
	SeedPub:         "seedpub",
	SeedPrv:         "seedprv",
	Normal:          "normal",
	DSGuard:         "dsguard",
}

var (
	lookUpTypes = []NodeType{Lookup, Seed, SeedPub, SeedPrv}
	nodeTypes   = []NodeType{Lookup, Seed, SeedPub, SeedPrv, Normal, DSGuard}
)

func (n NodeType) String() string {
	if s, ok := nodeTypeStringMap[n]; ok {
		return s
	}
	return ""
}

func NodeTypeFromString(typ string) NodeType {
	if t, ok := stringNodeTypeMap[strings.ToLower(typ)]; ok {
		return t
	}
	return UnknownNodeType
}

func IsGeneralLookup(nt NodeType) bool {
	var isLookup bool
	for _, typ := range lookUpTypes {
		if nt == typ {
			isLookup = true
		}
	}
	return isLookup
}
