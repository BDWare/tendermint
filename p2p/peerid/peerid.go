package peerid

import "github.com/tendermint/tendermint/crypto"

// For Tendermint P2P: ID is a hex-encoded crypto.Address
// For libp2p: ID is a base58 encoded libp2p ID string (ID.Pretty() of github.com/libp2p/go-libp2p-core/peer)
type ID string

// IDByteLength is the length of a crypto.Address. Currently only 20.
// Only used in Tendermint P2P ID for now.
// TODO: support other length addresses ?
const IDByteLength = crypto.AddressSize
