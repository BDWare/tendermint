package evidence

import (
	amino "github.com/tendermint/go-amino"

	cryptoamino "github.com/bdware/tendermint/crypto/encoding/amino"
	"github.com/bdware/tendermint/types"
)

var cdc = amino.NewCodec()

func init() {
	RegisterMessages(cdc)
	cryptoamino.RegisterAmino(cdc)
	types.RegisterEvidences(cdc)
}

// For testing purposes only
func RegisterMockEvidences() {
	types.RegisterMockEvidences(cdc)
}
