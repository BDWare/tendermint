package p2p

import (
	"github.com/bdware/tendermint/p2p/conn"
	peer2 "github.com/libp2p/go-libp2p-core/peer"
)

type ChannelDescriptor = conn.ChannelDescriptor
type ConnectionStatus = conn.ConnectionStatus

type Libp2pID = peer2.ID
