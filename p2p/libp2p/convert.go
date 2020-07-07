package libp2p

import (
	"net"
	"strconv"
	"strings"

	lppeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/bdware/tendermint/p2p"
)

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

func Multiaddr2NetAddr(id lppeer.ID, ma multiaddr.Multiaddr) *p2p.NetAddress {
	s := ma.String() // for example "/ip4/127.0.0.1/udp/1234"
	//s = s[5:]
	parts := strings.Split(s, "/")
	port, _ := strconv.Atoi(parts[4])
	return &p2p.NetAddress{
		ID:   LpID2ID(id),
		Port: uint16(port),
		IP:   net.ParseIP(parts[2]),
	}
}

func Multiaddr2DialString(ma multiaddr.Multiaddr) string {
	s := ma.String() // for example "/ip4/127.0.0.1/udp/1234"
	//s = s[5:]
	parts := strings.Split(s, "/")
	return parts[2] + ":" + parts[4]
}

func ID2lpID(id p2p.ID) p2p.Libp2pID {
	// now p2p.ID is human-readable base58 encoded string
	id1, err := lppeer.Decode(string(id))
	if err != nil {
		panic(err)
	}
	return id1
}

// In fact, libp2p peer.ID is not a string
func LpID2ID(id p2p.Libp2pID) p2p.ID {
	return p2p.ID(id.String())
}
