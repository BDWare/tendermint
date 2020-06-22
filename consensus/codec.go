package consensus

import (
	amino "github.com/tendermint/go-amino"

	"github.com/bdware/tendermint/types"
)

var cdc = amino.NewCodec()

func init() {
	RegisterMessages(cdc)
	RegisterWALMessages(cdc)
	types.RegisterBlockAmino(cdc)
}
