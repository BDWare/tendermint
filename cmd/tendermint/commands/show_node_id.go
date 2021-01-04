package commands

import (
	"fmt"
	"github.com/tendermint/tendermint/p2p/libp2p"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/p2p"
)

// ShowNodeIDCmd dumps node's ID to the standard output.
var ShowNodeIDCmd = &cobra.Command{
	Use:   "show_node_id",
	Short: "Show this node's ID",
	RunE:  showNodeID,
}

func showNodeID(cmd *cobra.Command, args []string) error {
	// TODO: refactor the following to a common function used by init/gen_node_key/show_node_id/testnet
	// like p2p.LoadNodeKey(path string, isLibp2p bool)
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
