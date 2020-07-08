package conn

import (
	amino "github.com/tendermint/go-amino"

	cryptoamino "github.com/bdware/tendermint/crypto/encoding/amino"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	cryptoamino.RegisterAmino(cdc)
	RegisterPacket(cdc)
	RegisterLibp2pPacket(cdc)
}
