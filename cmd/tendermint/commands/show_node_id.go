package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bdware/tendermint/p2p"
	"github.com/bdware/tendermint/p2p/libp2p"
)

// ShowNodeIDCmd dumps node's ID to the standard output.
var ShowNodeIDCmd = &cobra.Command{
	Use:   "show_node_id",
	Short: "Show this node's ID",
	RunE:  showNodeID,
}

func showNodeID(cmd *cobra.Command, args []string) error {
	var (
		nodeKey *p2p.NodeKey
		err     error
	)
	if !config.P2P.Libp2p {
		nodeKey, err = p2p.LoadNodeKey(config.NodeKeyFile())
	} else {
		nodeKey, err = libp2p.LoadLpNodeKey(config.NodeKeyFile())
	}
	if err != nil {
		return err
	}

	fmt.Println(nodeKey.ID())
	return nil
}
