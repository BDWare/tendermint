package p2p

import (
	"fmt"
	"github.com/bdware/tendermint/crypto"
	"github.com/bdware/tendermint/crypto/tmhash"
	tmos "github.com/bdware/tendermint/libs/os"
	libp2p_crypto "github.com/libp2p/go-libp2p-core/crypto"
	"io/ioutil"
)

//-----------libp2p helper functions used in tdm cmd like init, show_node_id------------
func LoadOrGenLpNodeKey(filePath string) (*NodeKey, error) {
	if tmos.FileExists(filePath) {
		nodeKey, err := LoadLpNodeKey(filePath)
		if err != nil {
			return nil, err
		}
		return nodeKey, nil
	}
	return genLpNodeKey(filePath)
}

func LoadLpNodeKey(filePath string) (*NodeKey, error) {
	keyBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	privKey, err := libp2p_crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading NodeKey from %v: %v", filePath, err)
	}
	nodeKey := &NodeKey{
		PrivKey: lpPrivKey{K: privKey},
	}
	return nodeKey, nil
}

func genLpNodeKey(filePath string) (*NodeKey, error) {
	privKey, _, _ := libp2p_crypto.GenerateEd25519Key(crypto.CReader())
	nodeKey := &NodeKey{
		PrivKey: lpPrivKey{K: privKey},
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

func GetNodeKeyFromLpPrivKey(pk libp2p_crypto.PrivKey) *NodeKey {
	return &NodeKey{
		PrivKey: lpPrivKey{K: pk},
	}
}

//----------------------libp2p priKey-------------------------
type lpPrivKey struct {
	K libp2p_crypto.PrivKey
}

func (l lpPrivKey) Bytes() []byte {
	bs, _ := l.K.Bytes()
	return bs
}

func (l lpPrivKey) Sign(msg []byte) ([]byte, error) {
	return l.K.Sign(msg)
}

func (l lpPrivKey) PubKey() crypto.PubKey {
	return lpPubKey{
		l.K.GetPublic(),
	}
}

func (l lpPrivKey) Equals(key crypto.PrivKey) bool {
	other, ok := key.(lpPrivKey)
	if !ok {
		return false
	}
	return l.K.Equals(other.K)
}

//----------------------libp2p pubKey-------------------------
type lpPubKey struct {
	K libp2p_crypto.PubKey
}

func (l lpPubKey) Address() crypto.Address {
	bs, _ := l.K.Raw()
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
