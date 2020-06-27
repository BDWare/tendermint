package p2p

import (
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"net"
	"strconv"
	"strings"
)

func NetAddr2LpAddrInfo(address NetAddress) libp2pPeer.AddrInfo {
	// maybe false
	maddr, err := multiaddr.NewMultiaddr(address.DialString())
	if err != nil {

	}
	return libp2pPeer.AddrInfo{
		Addrs: []multiaddr.Multiaddr{maddr},
		ID:    ID2lpID(address.ID),
	}
}

func NetAddr2Multiaddr(na NetAddress) multiaddr.Multiaddr {
	str := "/ip4/" + na.IP.String() + "/tcp/" + strconv.FormatUint(uint64(na.Port), 10)
	ma, _ := multiaddr.NewMultiaddr(str)
	return ma
}

func Multiaddr2NetAddr(id libp2pPeer.ID, ma multiaddr.Multiaddr)  *NetAddress{
	s := ma.String() // for example "/ip4/127.0.0.1/udp/1234"
	//s = s[5:]
	parts := strings.Split(s, "/")
	port, _ := strconv.Atoi(parts[4])
	return &NetAddress{
		ID: lpID2ID(id),
		Port: uint16(port),
		IP: net.ParseIP(parts[2]),
	}
}

func Multiaddr2DialString(ma multiaddr.Multiaddr) string{
	s := ma.String() // for example "/ip4/127.0.0.1/udp/1234"
	//s = s[5:]
	parts := strings.Split(s, "/")
	return parts[2] + ":" + parts[4]
}

func ID2lpID(id ID) libp2pPeer.ID{
	return libp2pPeer.ID(id)
}

func lpID2ID(id libp2pPeer.ID) ID {
	//return ID(id)
	return ID(id.String())
}
