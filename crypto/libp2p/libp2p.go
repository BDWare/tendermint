package libp2p

import (
	lpcrypto "github.com/libp2p/go-libp2p-core/crypto"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

const (
	PrivKeyName = "tendermint/libp2pPrivKey"
	PubKeyName  = "tendermint/libp2pPubKey"

	KeyType     = "libp2pKey"
)


func init() {
	tmjson.RegisterType(PubKey{}, PubKeyName)
	tmjson.RegisterType(PrivKey{}, PrivKeyName)
}

var _ crypto.PrivKey = PrivKey{}

// TODO: If use these keys in tendermint, must register these and some libp2p keys to amino first.
type PrivKey struct {
	K lpcrypto.PrivKey
}

func (k PrivKey) Type() string {
	return KeyType
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

var _ crypto.PubKey = PubKey{}

type PubKey struct {
	K lpcrypto.PubKey
}

func (l PubKey) Type() string {
	return KeyType
}

func (l PubKey) Address() crypto.Address {
	bs, _ := l.K.Raw()
	return crypto.Address(tmhash.SumTruncated(bs))
}

func (l PubKey) Bytes() []byte {
	bs, _ := l.K.Bytes()
	return bs
}

func (l PubKey) VerifySignature(msg []byte, sig []byte) bool {
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
