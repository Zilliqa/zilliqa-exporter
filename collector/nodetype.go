package collector

import "strings"

type NodeType int

const (
	Lookup NodeType = iota
	SeedAPIPub
	SeedAPIPrv
	Normal
	DSGuard
	UnknownNodeType
)

const (
	Level2Lookup = SeedAPIPub
	NewLookup    = SeedAPIPrv
)

var stringNodeTypeMap = map[string]NodeType{
	"lookup":       Lookup,
	"seed-apipub":  SeedAPIPub,
	"level2lookup": SeedAPIPub,
	"seed-apiprv":  SeedAPIPrv,
	"newlookup":    SeedAPIPrv,
	"normal":       Normal,
	"dsguard":      DSGuard,
	"":             UnknownNodeType,
}

var nodeTypeStringMap = map[NodeType]string{
	Lookup:          "lookup",
	SeedAPIPrv:      "seed-apiprv",
	SeedAPIPub:      "seed-apipub",
	Normal:          "normal",
	DSGuard:         "dsguard",
	UnknownNodeType: "",
}

var (
	lookUpTypes = []NodeType{Lookup, SeedAPIPub, SeedAPIPrv}
	nodeTypes   = []NodeType{Lookup, SeedAPIPub, SeedAPIPrv, Normal, DSGuard}
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
