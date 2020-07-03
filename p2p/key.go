package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bdware/tendermint/crypto/ed25519"
	libp2p_peer "github.com/libp2p/go-libp2p-core/peer"
	"io/ioutil"

	"github.com/bdware/tendermint/crypto"
	tmos "github.com/bdware/tendermint/libs/os"
)

// For temdermint p2p: ID is a hex-encoded crypto.Address
// For libp2p: ID is a base58 encoded libp2p ID string (ID.Pretty() of github.com/libp2p/go-libp2p-core/peer)
type ID string

//type ID interface {
//	isID()
//}
//type id string
//func (id id) isID() {}

// IDByteLength is only used in tendermint p2p ID for now.
// IDByteLength is the length of a crypto.Address. Currently only 20.
// TODO: support other length addresses ?
const IDByteLength = crypto.AddressSize

//------------------------------------------------------------------------------
// Persistent peer ID
// TODO: encrypt on disk

// NodeKey is the persistent peer key.
// It contains the nodes private key for authentication.
type NodeKey struct {
	PrivKey crypto.PrivKey `json:"priv_key"` // our priv key
}

// ID returns the peer's canonical ID - the hash of its public key.
func (nodeKey *NodeKey) ID() ID {
	return PubKeyToID(nodeKey.PubKey())
}

// PubKey returns the peer's PubKey
func (nodeKey *NodeKey) PubKey() crypto.PubKey {
	return nodeKey.PrivKey.PubKey()
}

// PubKeyToID returns the ID corresponding to the given PubKey.
// It's the hex-encoding of the pubKey.Address().
//func PubKeyToID(pubKey crypto.PubKey) ID {
//	return ID(hex.EncodeToString(pubKey.Address()))
//}

// PubKeyToID returns libp2p peer.ID or tendermint p2p peer ID
func PubKeyToID(pubKey crypto.PubKey) ID {
	// TODO: maybe change the way to get type of key
	// so that we can put lpKey in package libp2p
	if pk, ok := pubKey.(lpPubKey); ok {
		peerID, _ := libp2p_peer.IDFromPublicKey(pk.K)
		return LpID2ID(peerID)
	} else {
		return ID(hex.EncodeToString(pubKey.Address()))
	}
}

// LoadOrGenNodeKey attempts to load the NodeKey from the given filePath.
// If the file does not exist, it generates and saves a new NodeKey.
func LoadOrGenNodeKey(filePath string) (*NodeKey, error) {
	if tmos.FileExists(filePath) {
		nodeKey, err := LoadNodeKey(filePath)
		if err != nil {
			return nil, err
		}
		return nodeKey, nil
	}
	return genNodeKey(filePath)
}

func LoadNodeKey(filePath string) (*NodeKey, error) {
	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	nodeKey := new(NodeKey)
	err = Cdc.UnmarshalJSON(jsonBytes, nodeKey)
	if err != nil {
		return nil, fmt.Errorf("error reading NodeKey from %v: %v", filePath, err)
	}
	return nodeKey, nil
}

func genNodeKey(filePath string) (*NodeKey, error) {
	privKey := ed25519.GenPrivKey()
	nodeKey := &NodeKey{
		PrivKey: privKey,
	}

	jsonBytes, err := Cdc.MarshalJSON(nodeKey)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filePath, jsonBytes, 0600)
	if err != nil {
		return nil, err
	}
	return nodeKey, nil
}

//------------------------------------------------------------------------------

// MakePoWTarget returns the big-endian encoding of 2^(targetBits - difficulty) - 1.
// It can be used as a Proof of Work target.
// NOTE: targetBits must be a multiple of 8 and difficulty must be less than targetBits.
func MakePoWTarget(difficulty, targetBits uint) []byte {
	if targetBits%8 != 0 {
		panic(fmt.Sprintf("targetBits (%d) not a multiple of 8", targetBits))
	}
	if difficulty >= targetBits {
		panic(fmt.Sprintf("difficulty (%d) >= targetBits (%d)", difficulty, targetBits))
	}
	targetBytes := targetBits / 8
	zeroPrefixLen := (int(difficulty) / 8)
	prefix := bytes.Repeat([]byte{0}, zeroPrefixLen)
	mod := (difficulty % 8)
	if mod > 0 {
		nonZeroPrefix := byte(1<<(8-mod) - 1)
		prefix = append(prefix, nonZeroPrefix)
	}
	tailLen := int(targetBytes) - len(prefix)
	return append(prefix, bytes.Repeat([]byte{0xFF}, tailLen)...)
}
