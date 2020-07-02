package p2p

import (
	amino "github.com/tendermint/go-amino"

	cryptoamino "github.com/bdware/tendermint/crypto/encoding/amino"
)

var Cdc = amino.NewCodec()

func init() {
	cryptoamino.RegisterAmino(Cdc)
}
