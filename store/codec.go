package store

import (
	amino "github.com/tendermint/go-amino"

	"github.com/bdware/tendermint/types"
)

var cdc = amino.NewCodec()

func init() {
	types.RegisterBlockAmino(cdc)
}
