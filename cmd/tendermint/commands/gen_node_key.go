package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	tmos "github.com/bdware/tendermint/libs/os"
	"github.com/bdware/tendermint/p2p"
)

// GenNodeKeyCmd allows the generation of a node key. It prints node's ID to
// the standard output.
var GenNodeKeyCmd = &cobra.Command{
	Use:   "gen_node_key",
	Short: "Generate a node key for this node and print its ID",
	RunE:  genNodeKey,
}

func genNodeKey(cmd *cobra.Command, args []string) error {
	nodeKeyFile := config.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		return fmt.Errorf("node key at %s already exists", nodeKeyFile)
	}

	var (
		nodeKey *p2p.NodeKey
		err     error
	)
	if !config.P2P.Libp2p {
		nodeKey, err = p2p.LoadOrGenNodeKey(nodeKeyFile)
	} else {
		nodeKey, err = p2p.LoadOrGenLpNodeKey(nodeKeyFile)
	}
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}
