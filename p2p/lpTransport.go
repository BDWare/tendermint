package p2p

import (
	"context"
	"fmt"
	"github.com/bdware/tendermint/libs/cmap"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	ShakehandProtocol = "tdm-shakehand"

)


// ConnFilterFunc to be implemented by filter hooks after a new connection has
// been established. The set of exisiting connections is passed along together
// with all resolved IPs for the new connection.
//type ConnFilterFunc func(ConnSet, net.Conn, []net.IP) error

// ConnDuplicateIPFilter resolves and keeps all ips for an incoming connection
// and refuses new ones if they come from a known ip.
//func ConnDuplicateIPFilter() ConnFilterFunc {
//	return func(cs ConnSet, c net.Conn, ips []net.IP) error {
//		for _, ip := range ips {
//			if cs.HasIP(ip) {
//				return ErrRejected{
//					conn:        c,
//					err:         fmt.Errorf("ip<%v> already connected", ip),
//					isDuplicate: true,
//				}
//			}
//		}
//
//		return nil
//	}
//}

// MultiplexTransportOption sets an optional parameter on the
// LpTransport.
//type MultiplexTransportOption func(*LpTransport)
//
//// MultiplexTransportConnFilters sets the filters for rejection new connections.
//func MultiplexTransportConnFilters(
//	filters ...ConnFilterFunc,
//) MultiplexTransportOption {
//	return func(mt *LpTransport) { mt.connFilters = filters }
//}
//
//// MultiplexTransportFilterTimeout sets the timeout waited for filter calls to
//// return.
//func MultiplexTransportFilterTimeout(
//	timeout time.Duration,
//) MultiplexTransportOption {
//	return func(mt *LpTransport) { mt.filterTimeout = timeout }
//}
//
//// MultiplexTransportResolver sets the Resolver used for ip lokkups, defaults to
//// net.DefaultResolver.
//func MultiplexTransportResolver(resolver IPResolver) MultiplexTransportOption {
//	return func(mt *LpTransport) { mt.resolver = resolver }
//}
//
//// MultiplexTransportMaxIncomingConnections sets the maximum number of
//// simultaneous connections (incoming). Default: 0 (unlimited)
//func MultiplexTransportMaxIncomingConnections(n int) MultiplexTransportOption {
//	return func(mt *LpTransport) { mt.maxIncomingConnections = n }
//}

// LpTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type LpTransport struct {
	netAddr                NetAddress
	maxIncomingConnections int // see MaxIncomingConnections

	acceptc chan accept
	closec  chan struct{}

	// Lookup table for duplicate ip and id checks.
	//conns       ConnSet
	connFilters []ConnFilterFunc

	dialTimeout      time.Duration
	filterTimeout    time.Duration
	handshakeTimeout time.Duration
	nodeInfo         NodeInfo
	nodeKey          NodeKey

	wait4Peer *cmap.CMap
	host host.Host
}

// Test multiplexTransport for interface completeness.
var _ Transport = (*LpTransport)(nil)
var _ transportLifecycle = (*LpTransport)(nil)

// NewLpTransport returns a tcp connected multiplexed lpPeer.
func NewLpTransport(nodeInfo NodeInfo, nodeKey NodeKey, host host.Host, ) *LpTransport {
	mt := &LpTransport{
		acceptc:          make(chan accept),
		closec:           make(chan struct{}),
		dialTimeout:      defaultDialTimeout,
		filterTimeout:    defaultFilterTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,
		host: host,
		wait4Peer: cmap.NewCMap(),
	}
	// set our address (used in switch)
	addr, err := nodeInfo.NetAddress()
	if err != nil {
		panic(err)
	}
	mt.netAddr = *addr
	go mt.handleConn()
	mt.host.SetStreamHandler(ShakehandProtocol, func(s network.Stream) {
		prID := s.Conn().LocalPeer()
		nodeInfo, err := mt.shakehand(s)
		if err != nil {
			//mt.host.Network().ClosePeer(prID)
			return
		}
		ma := s.Conn().RemoteMultiaddr()
		select {
		case mt.acceptc <- accept{nodeInfo: nodeInfo, netAddr: Multiaddr2NetAddr(prID, ma)}:

		case <-mt.closec:

		}
	})
	return mt
}

// NetAddress implements Transport.
func (mt *LpTransport) NetAddress() NetAddress {
	return mt.netAddr
}

func (mt *LpTransport) handleConn()  {
	sub, err := mt.host.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))
	if err != nil {

	}
	for evt := range sub.Out() {
		e, ok := evt.(event.EvtPeerConnectednessChanged)
		if !ok {
			continue
		}
		mt.handleConnRoutine(e)
	}

}

func (mt *LpTransport) handleConnRoutine(e event.EvtPeerConnectednessChanged) {
	prID := e.Peer
	id := lpID2ID(prID)
	if e.Connectedness == network.Connected {
		c := mt.host.Network().ConnsToPeer(prID)[0]
		if c.Stat().Direction == network.DirOutbound {
			// the peer that starts the connection also inits the handshake
			s, err := mt.host.NewStream(context.TODO(), prID, ShakehandProtocol)
			if err != nil {
				//mt.host.Network().ClosePeer(prID)
				return
			}

			nodeInfo, err := mt.shakehand(s)
			if err != nil {
				//mt.host.Network().ClosePeer(prID)
				return
			}
			ma := c.RemoteMultiaddr()
			if mt.wait4Peer.Has(string(id)) {
				ch := mt.wait4Peer.Get(string(id)).(chan accept)
				ch <- accept{nodeInfo: nodeInfo, netAddr: Multiaddr2NetAddr(prID, ma)}
			} else {
				select {
				case mt.acceptc <- accept{nodeInfo: nodeInfo, netAddr: Multiaddr2NetAddr(prID, ma)}:

				case <-mt.closec:

				}
			}
		}
	}
}

// Accept implements Transport.
func (mt *LpTransport) Accept(cfg peerConfig) (Peer, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, a.err
		}

		cfg.outbound = false

		return mt.wrapLpPeer(a.nodeInfo, cfg, a.netAddr), nil
	case <-mt.closec:
		return nil, ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *LpTransport) Dial(
	addr NetAddress,
	cfg peerConfig,
) (Peer, error) {
	ch := make(chan accept)
	mt.wait4Peer.Set(string(addr.ID), ch)
	ctx, _ := context.WithTimeout(context.Background(), mt.dialTimeout)
	ai := addr.LpAddrInfo()
	err := mt.host.Connect(ctx, ai)
	if err != nil {
		return nil, err
	}
	a := <- ch // block until connected event is detected
	mt.wait4Peer.Delete(string(addr.ID))

	cfg.outbound = true
	p := mt.wrapLpPeer(a.nodeInfo, cfg, &addr)

	return p, nil
}

// Close implements transportLifecycle.
func (mt *LpTransport) Close() error {
	close(mt.closec)

	return nil
}

// Listen implements transportLifecycle.
func (mt *LpTransport) Listen(addr NetAddress) (err error) {
	//ma := addr.Multiaddr()
	//if err = mt.host.Network().Listen(ma); err != nil {return err}
	//mt.netAddr = addr
	//mt.host.SetStreamHandler(ShakehandProtocol, func(s network.Stream) {
	//	prID := s.Conn().LocalPeer()
	//	nodeInfo, err := mt.shakehand(s)
	//	if err != nil {
	//		mt.host.Network().ClosePeer(prID)
	//	}
	//	ma := s.Conn().RemoteMultiaddr()
	//	select {
	//	case mt.acceptc <- accept{nodeInfo: nodeInfo, netAddr: Multiaddr2NetAddr(prID, ma)}:
	//
	//	case <-mt.closec:
	//
	//	}
	//})
	return fmt.Errorf("should not be called")
}

// shakehand exchanges and checks NodeInfo
func (mt *LpTransport) shakehand(s network.Stream) (NodeInfo, error){
	var (
		errc = make(chan error, 2)

		peerNodeInfo DefaultNodeInfo
		ourNodeInfo  = mt.nodeInfo.(DefaultNodeInfo)
	)

	go func(errc chan<- error, s network.Stream) {
		_, err := cdc.MarshalBinaryLengthPrefixedWriter(s, ourNodeInfo)
		errc <- err
	}(errc, s)

	go func(errc chan<- error, s network.Stream) {
		_, err := cdc.UnmarshalBinaryLengthPrefixedReader(
			s,
			&peerNodeInfo,
			int64(MaxNodeInfoSize()),
		)
		errc <- err
	}(errc, s)

	for i := 0; i < cap(errc); i++ {
		err := <-errc
		if err != nil {
			return nil, err
		}
	}

	if err := peerNodeInfo.Validate(); err != nil {
		return nil, err
	}
	lpID, _ := libp2pPeer.IDFromPublicKey(s.Conn().RemotePublicKey())
	connID := lpID2ID(lpID)
	// Ensure connection key matches self reported key.
	if connID != peerNodeInfo.ID() {
		return nil, fmt.Errorf(
				"conn.ID (%v) NodeInfo.ID (%v) mismatch",
				connID,
				peerNodeInfo.ID(),
			)
	}

	// Reject self.
	if mt.nodeInfo.ID() == peerNodeInfo.ID() {
		return nil, fmt.Errorf("reject self")
	}

	if err := mt.nodeInfo.CompatibleWith(peerNodeInfo); err != nil {
		return nil, fmt.Errorf("not compatible")
	}


	return peerNodeInfo, nil
}


// Cleanup removes the given address from the connections set and
// closes the connection.
func (mt *LpTransport) Cleanup(p Peer) {
	_ = p.CloseConn()
}

func (mt *LpTransport) wrapLpPeer(
	ni NodeInfo,
	cfg peerConfig,
	socketAddr *NetAddress,
) Peer {

	persistent := false
	if cfg.isPersistent != nil {
		if cfg.outbound {
			persistent = cfg.isPersistent(socketAddr)
		} else {
			selfReportedAddr, err := ni.NetAddress()
			if err == nil {
				persistent = cfg.isPersistent(selfReportedAddr)
			}
		}
	}

	p := newLpPeer(
		ni,
		cfg.reactorsByCh,
		cfg.chDescs,
		cfg.onPeerError,
		//PeerMetrics(cfg.metrics),
	)

	p.persistent = persistent
	p.outbound = cfg.outbound
	p.socketAddr = socketAddr

	return p
}
