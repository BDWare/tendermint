package libp2p

import (
	lpcrypto "github.com/libp2p/go-libp2p-core/crypto"

	"github.com/bdware/tendermint/crypto"
	"github.com/bdware/tendermint/crypto/tmhash"
)

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
