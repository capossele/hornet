package anchor

import (
	flag "github.com/spf13/pflag"

	"github.com/gohornet/hornet/pkg/node"
)

const (
	// the ID of the child.
	CfgAnchorChildID = "anchor.childID"
	// the URL of the parent API.
	CfgAnchorParentAPI = "anchor.parentAPI"
)

var params = &node.PluginParams{
	Params: map[string]*flag.FlagSet{
		"nodeConfig": func() *flag.FlagSet {
			fs := flag.NewFlagSet("", flag.ContinueOnError)
			fs.String(CfgAnchorChildID, "mainnet", "the string identifier of the child")
			fs.String(CfgAnchorParentAPI, "http://localhost:8070", "the bind address on which the parent can be accessed from")
			return fs
		}(),
	},
	Masked: nil,
}
