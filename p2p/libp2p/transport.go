package libp2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/multiformats/go-multiaddr"

	"github.com/bdware/tendermint/crypto"
	"github.com/bdware/tendermint/libs/cmap"
	"github.com/bdware/tendermint/p2p"
	"github.com/bdware/tendermint/p2p/conn"
	"github.com/bdware/tendermint/p2p/libp2p/util"
)

const (
	ShakehandProtocol = "tdm-handshake"
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
	//s        net.Conn
	nodeInfo p2p.NodeInfo
	err      error
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

	wait4Peer *cmap.CMap
	host      host.Host
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
		wait4Peer:        cmap.NewCMap(),
		host:             host,
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
		na := p2p.NewNetAddressLibp2pIDMultiaddr(prID, s.Conn().RemoteMultiaddr())
		nodeInfo, err := mt.dohandshake(s, nil)
		if err != nil {
			// If Close() has been called, silently exit.
			select {
			case _, ok := <-mt.closec:
				if !ok {
					return
				}
			default:
				// Transport is not closed
			}

			mt.acceptc <- accept{netAddr: na, err: err}
			return
		}

		select {
		case mt.acceptc <- accept{netAddr: na, nodeInfo: nodeInfo}:

		case <-mt.closec:

		}

		// don't need shakehand stream any longer
		s.Close()
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
		if c.Stat().Direction == network.DirOutbound {
			mt := n2.mt
			prID := c.RemotePeer()
			ID := util.Libp2pID2ID(prID)
			ma := c.RemoteMultiaddr()
			na := p2p.NewNetAddressLibp2pIDMultiaddr(prID, ma)

			// whether we are dialing this peer actively?
			dialing := mt.wait4Peer.Has(string(ID))

			// the peer that starts the connection also inits the handshake
			s, err := mt.host.NewStream(context.TODO(), prID, ShakehandProtocol)
			if err != nil {
				// Close Conn so that we may connect to this peer later
				c.Close()
				if dialing {
					mt.wait4Peer.Get(string(ID)).(chan accept) <- accept{netAddr: na, err: err}
				}
				return
			}

			nodeInfo, err := mt.dohandshake(s, nil)
			if err != nil {
				if dialing {
					mt.wait4Peer.Get(string(ID)).(chan accept) <- accept{netAddr: na, err: err}
				}
				return
			}
			// we don't need this shakehand stream any longer
			s.Close()

			if dialing {
				mt.wait4Peer.Get(string(ID)).(chan accept) <- accept{netAddr: na, nodeInfo: nodeInfo}
				return
			}

			// TODO: add outbound peer of dht discovery to switch

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

// Accept implements Transport.
func (mt *LpTransport) Accept(cfg p2p.PeerConfig) (p2p.Peer, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, a.err
		}

		cfg.Outbound = false

		return mt.wrapLpPeer(a.nodeInfo, cfg, a.netAddr), nil
	case <-mt.closec:
		return nil, p2p.ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *LpTransport) Dial(
	addr p2p.NetAddress,
	cfg p2p.PeerConfig,
) (p2p.Peer, error) {
	ch := make(chan accept)
	mt.wait4Peer.Set(string(addr.ID), ch)
	defer mt.wait4Peer.Delete(string(addr.ID))

	ctx, _ := context.WithTimeout(context.Background(), mt.dialTimeout)
	ai := p2p.LpAddrInfoFromNetAddress(addr)
	err := mt.host.Connect(ctx, ai)
	if err != nil {
		if err == swarm.ErrDialToSelf {
			return nil, p2p.NewIsSelfErrRejected(addr, err, addr.ID)
		}
		return nil, err
	}

	a := <-ch // block until connected event is detected
	if a.err != nil {
		return nil, a.err
	}

	cfg.Outbound = true
	p := mt.wrapLpPeer(a.nodeInfo, cfg, &addr)
	return p, nil
}

// Close implements transportLifecycle.
func (mt *LpTransport) Close() error {
	close(mt.closec)

	// closing of host should be done in Node.OnStop
	//if mt.host != nil {
	//	mt.host.Close()
	//}

	return nil
}

// Listen implements transportLifecycle.
func (mt *LpTransport) Listen(addr p2p.NetAddress) (err error) {
	//ma := addr.Multiaddr()
	//if err = mt.host.Network().Listen(ma); err != nil {return err}
	//mt.netAddr = addr
	//mt.host.SetStreamHandler(ShakehandProtocol, func(s network.Stream) {
	//	prID := s.Conn().LocalPeer()
	//	nodeInfo, err := mt.handshake(s)
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

// handshake exchanges NodeInfo
func handshake(
	s network.Stream,
	timeout time.Duration,
	nodeInfo p2p.NodeInfo,
) (p2p.NodeInfo, error) {
	if err := s.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	var (
		errc = make(chan error, 2)

		peerNodeInfo p2p.DefaultNodeInfo
		ourNodeInfo  = nodeInfo.(p2p.DefaultNodeInfo)
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

	return peerNodeInfo, s.SetDeadline(time.Time{})
}

// Cleanup removes the given address from the connections set and
// closes the connection.
func (mt *LpTransport) Cleanup(p p2p.Peer) {
	_ = p.CloseConn()
}

func (mt *LpTransport) cleanup(s network.Stream) error {
	return s.Conn().Close()
}

func (mt *LpTransport) dohandshake(
	s network.Stream,
	dialedAddr *p2p.NetAddress,
) (nodeInfo p2p.NodeInfo, err error) {
	defer func() {
		if err != nil {
			_ = mt.cleanup(s)
		}
	}()

	prID := s.Conn().RemotePeer()
	ID := util.Libp2pID2ID(prID)
	ma := s.Conn().RemoteMultiaddr()
	na := p2p.NewNetAddressLibp2pIDMultiaddr(prID, ma)

	//secretConn, err = upgradeSecretConn(s, mt.handshakeTimeout, mt.nodeKey.PrivKey)
	//if err != nil {
	//	err = fmt.Errorf("secret conn failed: %v", err)
	//	return nil, nil, p2p.NewIsAuthFailureErrRejected(*na, err, ID)
	//}

	// For outgoing conns, ensure connection key matches dialed key.
	connID := util.Libp2pID2ID(s.Conn().RemotePeer())
	if dialedAddr != nil {
		if dialedID := dialedAddr.ID; connID != dialedID {
			err = fmt.Errorf("conn.ID (%v) dialed ID (%v) mismatch", connID, dialedID)
			return nil, p2p.NewIsAuthFailureErrRejected(*na, err, ID)
		}
	}

	nodeInfo, err = handshake(s, mt.handshakeTimeout, mt.nodeInfo)
	if err != nil {
		err = fmt.Errorf("handshake failed: %v", err)
		return nil, p2p.NewIsAuthFailureErrRejected(*na, err, ID)
	}

	if err := nodeInfo.Validate(); err != nil {
		return nil, p2p.NewNodeInfoInvalidErrRejected(*na, err, ID)
	}

	// Ensure connection key matches self reported key.
	if connID != nodeInfo.ID() {
		err = fmt.Errorf("conn.ID (%v) NodeInfo.ID (%v) mismatch", connID, nodeInfo.ID())
		return nil, p2p.NewIsAuthFailureErrRejected(*na, err, ID)
	}

	// Reject self.
	if mt.nodeInfo.ID() == nodeInfo.ID() {
		return nil, p2p.NewIsSelfErrRejected(*na, err, ID)
	}

	if err := mt.nodeInfo.CompatibleWith(nodeInfo); err != nil {
		return nil, p2p.NewIncompatibleErrRejected(*na, err, ID)
	}

	return nodeInfo, nil
}

func upgradeSecretConn(
	s network.Stream,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := s.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(s, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}

func (mt *LpTransport) wrapLpPeer(
	ni p2p.NodeInfo,
	cfg p2p.PeerConfig,
	socketAddr *p2p.NetAddress,
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
		mt.host,
		ni,
		cfg.ReactorsByCh,
		cfg.ChDescs,
		cfg.OnPeerError,
		PeerMetrics(cfg.Metrics),
	)

	p.persistent = persistent
	p.outbound = cfg.Outbound
	p.socketAddr = socketAddr
	return p
}
