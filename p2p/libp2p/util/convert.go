package util

import (
	"encoding/hex"
	"strings"

	lppeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/tendermint/tendermint/crypto"
	lpkey "github.com/tendermint/tendermint/crypto/libp2p"
	"github.com/tendermint/tendermint/p2p/peerid"
)

func PubKeyToID(pubKey crypto.PubKey) peerid.ID {
	if pk, ok := pubKey.(lpkey.PubKey); ok {
		peerID, _ := lppeer.IDFromPublicKey(pk.K)
		return Libp2pID2ID(peerID)
	} else {
		return peerid.ID(hex.EncodeToString(pubKey.Address()))
	}
}

//func NetAddr2LpAddrInfo(address NetAddress) lppeer.AddrInfo {
//	// maybe false
//	maddr, err := multiaddr.NewMultiaddr(address.DialString())
//	if err != nil {
//
//	}
//	return lppeer.AddrInfo{
//		Addrs: []multiaddr.Multiaddr{maddr},
//		ID:    ID2lpID(address.ID),
//	}
//}
//
//func NetAddr2Multiaddr(na NetAddress) multiaddr.Multiaddr {
//	str := "/ip4/" + na.IP.String() + "/tcp/" + strconv.FormatUint(uint64(na.Port), 10)
//	ma, _ := multiaddr.NewMultiaddr(str)
//	return ma
//}

func Multiaddr2DialString(ma multiaddr.Multiaddr) string {
	s := ma.String() // for example "/ip4/127.0.0.1/udp/1234"
	//s = s[5:]
	parts := strings.Split(s, "/")
	return parts[2] + ":" + parts[4]
}

func ID2Libp2pID(id peerid.ID) lppeer.ID {
	// now p2p.ID is human-readable base58 encoded string
	id1, err := lppeer.Decode(string(id))
	if err != nil {
		panic(err)
	}
	return id1
}

// In fact, libp2p peer.ID is not a string
func Libp2pID2ID(id lppeer.ID) peerid.ID {
	return peerid.ID(id.String())
}
