package libp2p

import (
	"fmt"
	"io/ioutil"

	lpcrypto "github.com/libp2p/go-libp2p-core/crypto"

	"github.com/tendermint/tendermint/crypto"
	lpkey "github.com/tendermint/tendermint/crypto/libp2p"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/p2p"
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
		PrivKey: lpkey.PrivKey{K: privKey},
	}
	return nodeKey, nil
}

func genLpNodeKey(filePath string) (*p2p.NodeKey, error) {
	privKey, _, _ := lpcrypto.GenerateEd25519Key(crypto.CReader())
	nodeKey := &p2p.NodeKey{
		PrivKey: lpkey.PrivKey{K: privKey},
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
		PrivKey: lpkey.PrivKey{K: pk},
	}
}
