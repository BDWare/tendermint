package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/bdware/tendermint/crypto"
	"github.com/bdware/tendermint/crypto/tmhash"
	tmos "github.com/bdware/tendermint/libs/os"
	libp2p_crypto "github.com/libp2p/go-libp2p-core/crypto"
	libp2p_peer "github.com/libp2p/go-libp2p-core/peer"
)

// ID is a hex-encoded crypto.Address
type ID string

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

// PubKeyToID returns libp2p peer.ID
func PubKeyToID(pubKey crypto.PubKey) ID {
	if pk, ok := pubKey.(lpPubKey); ok {
		peerID, _ := libp2p_peer.IDFromPublicKey(pk.K)
		return lpID2ID(peerID)
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

// TODO: need modified
func LoadNodeKey(filePath string) (*NodeKey, error) {
	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	nodeKey := new(NodeKey)
	err = cdc.UnmarshalJSON(jsonBytes, nodeKey)
	if err != nil {
		return nil, fmt.Errorf("error reading NodeKey from %v: %v", filePath, err)
	}
	return nodeKey, nil
}

//func genNodeKey(filePath string) (*NodeKey, error) {
//	privKey := ed25519.GenPrivKey()
//	nodeKey := &NodeKey{
//		PrivKey: privKey,
//	}
//
//	jsonBytes, err := cdc.MarshalJSON(nodeKey)
//	if err != nil {
//		return nil, err
//	}
//	err = ioutil.WriteFile(filePath, jsonBytes, 0600)
//	if err != nil {
//		return nil, err
//	}
//	return nodeKey, nil
//}

func genNodeKey(filePath string) (*NodeKey, error) {
	privKey, _, _ := libp2p_crypto.GenerateEd25519Key(crypto.CReader())
	//privKey := ed25519.GenPrivKey()
	nodeKey := &NodeKey{
		PrivKey: LpPrivKey{K: privKey},
	}

	// TODO: change to human readable?
	keyBytes, err := privKey.Bytes()
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filePath, keyBytes, 0600)
	if err != nil {
		return nil, err
	}
	return nodeKey, nil
}

func GetNodeKeyFromLpPrivKey(pk libp2p_crypto.PrivKey) *NodeKey{
	return &NodeKey{
		PrivKey: LpPrivKey{K: pk},
	}
}

type LpPrivKey struct {
	K libp2p_crypto.PrivKey
}

func (l LpPrivKey) Bytes() []byte {
	bs, _ := l.K.Bytes()
	return bs
}

func (l LpPrivKey) Sign(msg []byte) ([]byte, error) {
	return l.K.Sign(msg)
}

func (l LpPrivKey) PubKey() crypto.PubKey {
	return lpPubKey{
		l.K.GetPublic(),
	}
}

func (l LpPrivKey) Equals(key crypto.PrivKey) bool {
	other, ok := key.(LpPrivKey)
	if !ok {
		return false
	}
	return l.K.Equals(other.K)
}

type lpPubKey struct {
	K libp2p_crypto.PubKey
}

func (l lpPubKey) Address() crypto.Address {
	bs,_ := l.K.Raw()
	return crypto.Address(tmhash.SumTruncated(bs))
}

func (l lpPubKey) Bytes() []byte {
	bs, _ := l.K.Bytes()
	return bs
}

func (l lpPubKey) VerifyBytes(msg []byte, sig []byte) bool {
	ok, _ := l.K.Verify(msg, sig)
	return ok
}

func (l lpPubKey) Equals(key crypto.PubKey) bool {
	other, ok := key.(lpPubKey)
	if !ok {
		return false
	}
	return l.K.Equals(other.K)
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
