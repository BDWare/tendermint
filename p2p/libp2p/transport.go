package libp2p

import (
	"context"
	"fmt"
	"github.com/bdware/tendermint/p2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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

type accept struct {
	netAddr  *p2p.NetAddress
	s        network.Stream
	nodeInfo p2p.NodeInfo
	err      error
	outbound bool
}

// LpTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type LpTransport struct {
	netAddr                p2p.NetAddress
	maxIncomingConnections int // see MaxIncomingConnections

	acceptc chan accept
	closec  chan struct{}

	// Lookup table for duplicate ip and id checks.
	//conns       ConnSet
	connFilters []p2p.ConnFilterFunc

	dialTimeout      time.Duration
	filterTimeout    time.Duration
	handshakeTimeout time.Duration
	nodeInfo         p2p.NodeInfo
	nodeKey          p2p.NodeKey

	//wait4Peer *cmap.CMap
	host host.Host
}

// Test multiplexTransport for interface completeness.
var _ p2p.Transport = (*LpTransport)(nil)
var _ p2p.TransportLifecycle = (*LpTransport)(nil)


// NewLpTransport returns a tcp connected multiplexed peer.
func NewLpTransport(nodeInfo p2p.NodeInfo, nodeKey p2p.NodeKey, host host.Host) *LpTransport {
	mt := &LpTransport{
		acceptc:          make(chan accept),
		closec:           make(chan struct{}),
		dialTimeout:      p2p.DefaultDialTimeout,
		filterTimeout:    p2p.DefaultFilterTimeout,
		handshakeTimeout: p2p.DefaultHandshakeTimeout,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,
		host:             host,
		//wait4Peer:        cmap.NewCMap(),
	}

	// set our address (used in switch)
	addr, err := nodeInfo.NetAddress()
	if err != nil {
		panic(err)
	}
	mt.netAddr = *addr

	mt.host.Network().Notify(&notif{mt: mt})
	mt.host.SetStreamHandler(ShakehandProtocol, func(s network.Stream) {
		prID := s.Conn().RemotePeer()
		nodeInfo, err := mt.shakehand(s)
		if err != nil {
			//mt.host.Network().ClosePeer(prID)
			return
		}
		ma := s.Conn().RemoteMultiaddr()
		select {
		case mt.acceptc <- accept{nodeInfo: nodeInfo, netAddr: p2p.Multiaddr2NetAddr(prID, ma), s: s, outbound: false}:

		case <-mt.closec:

		}
	})
	return mt
}

// NetAddress implements Transport.
func (mt *LpTransport) NetAddress() p2p.NetAddress {
	return mt.netAddr
}

type notif struct {
	mt *LpTransport
}

func (n2 *notif) Listen(n network.Network, m multiaddr.Multiaddr) {
	return
}

func (n2 *notif) ListenClose(n network.Network, m multiaddr.Multiaddr) {
	return
}

func (n2 *notif) Connected(n network.Network, c network.Conn) {
	// If don't run it in a go routine, Connect in Dial may not return directly and then timeout.
	go func() {
		mt := n2.mt
		prID := c.RemotePeer()
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

			select {
			case mt.acceptc <- accept{nodeInfo: nodeInfo, netAddr: p2p.Multiaddr2NetAddr(prID, ma), s: s, outbound: true}:

			case <-mt.closec:

			}
		}
	}()
}

func (n2 *notif) Disconnected(n network.Network, conn network.Conn) {
	return

}

func (n2 *notif) OpenedStream(n network.Network, stream network.Stream) {
	return

}

func (n2 *notif) ClosedStream(n network.Network, stream network.Stream) {
	return
}

//func (mt *LpTransport) handleConn() {
//	sub, err := mt.host.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))
//	if err != nil {
//
//	}
//	for evt := range sub.Out() {
//		e, ok := evt.(event.EvtPeerConnectednessChanged)
//		if !ok {
//			continue
//		}
//		mt.handleConnRoutine(e)
//	}
//
//}
//
//func (mt *LpTransport) handleConnRoutine(e event.EvtPeerConnectednessChanged) {
//	prID := e.Peer
//	id := lpID2ID(prID)
//	if e.Connectedness == network.Connected {
//		c := mt.host.Network().ConnsToPeer(prID)[0]
//		if c.Stat().Direction == network.DirOutbound {
//			// the peer that starts the connection also inits the handshake
//			s, err := mt.host.NewStream(context.TODO(), prID, ShakehandProtocol)
//			if err != nil {
//				//mt.host.Network().ClosePeer(prID)
//				return
//			}
//
//			nodeInfo, err := mt.shakehand(s)
//			if err != nil {
//				//mt.host.Network().ClosePeer(prID)
//				return
//			}
//			ma := c.RemoteMultiaddr()
//			if mt.wait4Peer.Has(string(id)) {
//				ch := mt.wait4Peer.Get(string(id)).(chan accept)
//				ch <- accept{nodeInfo: nodeInfo, netAddr: Multiaddr2NetAddr(prID, ma)}
//			} else {
//				select {
//				case mt.acceptc <- accept{nodeInfo: nodeInfo, netAddr: Multiaddr2NetAddr(prID, ma)}:
//
//				case <-mt.closec:
//
//				}
//			}
//		}
//	}
//}

// Accept implements Transport.
func (mt *LpTransport) Accept(cfg p2p.PeerConfig) (p2p.Peer, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, a.err
		}

		cfg.Outbound = a.outbound

		return mt.wrapLpPeer(a.nodeInfo, cfg, a.netAddr, a.s), nil
	case <-mt.closec:
		return nil, p2p.ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *LpTransport) Dial(
	addr p2p.NetAddress,
	cfg p2p.PeerConfig,
) (p2p.Peer, error) {
	//ch := make(chan accept)
	//mt.wait4Peer.Set(string(addr.ID), ch)
	//defer mt.wait4Peer.Delete(string(addr.ID))
	//
	//ctx, _ := context.WithTimeout(context.Background(), mt.dialTimeout)
	//ai := addr.LpAddrInfo()
	//err := mt.host.Connect(ctx, ai)
	//if err != nil {
	//	return nil, err
	//}
	//
	//a := <-ch // block until
	//ed event is detected
	//cfg.outbound = true
	//p := mt.wrapLpPeer(a.nodeInfo, cfg, &addr)
	//return p, nil
	return nil, fmt.Errorf("shouldn't be called")
}

// Close implements transportLifecycle.
func (mt *LpTransport) Close() error {
	close(mt.closec)

	if mt.host != nil {
		mt.host.Close()
	}

	return nil
}

// Listen implements transportLifecycle.
func (mt *LpTransport) Listen(addr p2p.NetAddress) (err error) {
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
func (mt *LpTransport) shakehand(s network.Stream) (p2p.NodeInfo, error) {
	var (
		errc = make(chan error, 2)

		peerNodeInfo p2p.DefaultNodeInfo
		ourNodeInfo  = mt.nodeInfo.(p2p.DefaultNodeInfo)
	)

	go func(errc chan<- error, s network.Stream) {
		_, err := p2p.Cdc.MarshalBinaryLengthPrefixedWriter(s, ourNodeInfo)
		errc <- err
	}(errc, s)

	go func(errc chan<- error, s network.Stream) {
		_, err := p2p.Cdc.UnmarshalBinaryLengthPrefixedReader(
			s,
			&peerNodeInfo,
			int64(p2p.MaxNodeInfoSize()),
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
	connID := p2p.LpID2ID(lpID)
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
func (mt *LpTransport) Cleanup(p p2p.Peer) {
	_ = p.CloseConn()
}

func (mt *LpTransport) wrapLpPeer(
	ni p2p.NodeInfo,
	cfg p2p.PeerConfig,
	socketAddr *p2p.NetAddress,
	s network.Stream,
) p2p.Peer {

	persistent := false
	if cfg.IsPersistent != nil {
		if cfg.Outbound {
			persistent = cfg.IsPersistent(socketAddr)
		} else {
			selfReportedAddr, err := ni.NetAddress()
			if err == nil {
				persistent = cfg.IsPersistent(selfReportedAddr)
			}
		}
	}

	p := newPeer(
		ni,
		cfg.ReactorsByCh,
		mt.host,
		cfg.ChDescs,
		cfg.OnPeerError,
		//PeerMetrics(cfg.metrics),
	)

	p.persistent = persistent
	p.outbound = cfg.Outbound
	p.socketAddr = socketAddr
	p.stream = s
	go p.recvRoutine()
	return p
}
