package libp2p

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	lpcrypto "github.com/libp2p/go-libp2p-core/crypto"
	lppeer "github.com/libp2p/go-libp2p-core/peer"

	"github.com/bdware/tendermint/crypto"
	"github.com/bdware/tendermint/crypto/tmhash"
	tmos "github.com/bdware/tendermint/libs/os"
	"github.com/bdware/tendermint/p2p"
)

//-----------libp2p helper functions used in tdm cmd like init, show_node_id------------
func LoadOrGenLpNodeKey(filePath string) (*p2p.NodeKey, error) {
	if tmos.FileExists(filePath) {
		nodeKey, err := LoadLpNodeKey(filePath)
		if err != nil {
			return nil, err
		}
		return nodeKey, nil
	}
	return genLpNodeKey(filePath)
}

func LoadLpNodeKey(filePath string) (*p2p.NodeKey, error) {
	keyBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	privKey, err := lpcrypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading NodeKey from %v: %v", filePath, err)
	}
	nodeKey := &p2p.NodeKey{
		PrivKey: PrivKey{K: privKey},
	}
	return nodeKey, nil
}

func genLpNodeKey(filePath string) (*p2p.NodeKey, error) {
	privKey, _, _ := lpcrypto.GenerateEd25519Key(crypto.CReader())
	nodeKey := &p2p.NodeKey{
		PrivKey: PrivKey{K: privKey},
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

func GetNodeKeyFromLpPrivKey(pk lpcrypto.PrivKey) *p2p.NodeKey {
	return &p2p.NodeKey{
		PrivKey: PrivKey{K: pk},
	}
}

//----------------------libp2p priKey-------------------------
type PrivKey struct {
	K lpcrypto.PrivKey
}

func (k PrivKey) Bytes() []byte {
	bs, _ := k.K.Bytes()
	return bs
}

func (k PrivKey) Sign(msg []byte) ([]byte, error) {
	return k.K.Sign(msg)
}

func (k PrivKey) PubKey() crypto.PubKey {
	return PubKey{
		k.K.GetPublic(),
	}
}

func (k PrivKey) Equals(key crypto.PrivKey) bool {
	other, ok := key.(PrivKey)
	if !ok {
		return false
	}
	return k.K.Equals(other.K)
}

//----------------------libp2p pubKey-------------------------
type PubKey struct {
	K lpcrypto.PubKey
}

func (l PubKey) Address() crypto.Address {
	bs, _ := l.K.Raw()
	return crypto.Address(tmhash.SumTruncated(bs))
}

func (l PubKey) Bytes() []byte {
	bs, _ := l.K.Bytes()
	return bs
}

func (l PubKey) VerifyBytes(msg []byte, sig []byte) bool {
	ok, _ := l.K.Verify(msg, sig)
	return ok
}

func (l PubKey) Equals(key crypto.PubKey) bool {
	other, ok := key.(PubKey)
	if !ok {
		return false
	}
	return l.K.Equals(other.K)
}

func PubKeyToID(pubKey crypto.PubKey) p2p.ID {
	// TODO: maybe change the way to get type of key
	// so that we can put lpKey in package libp2p
	if pk, ok := pubKey.(PubKey); ok {
		peerID, _ := lppeer.IDFromPublicKey(pk.K)
		return LpID2ID(peerID)
	} else {
		return p2p.ID(hex.EncodeToString(pubKey.Address()))
	}
}
