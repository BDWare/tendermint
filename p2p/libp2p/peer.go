package libp2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"net"
	"time"

	"github.com/libp2p/go-libp2p-core/host"

	"github.com/bdware/tendermint/libs/cmap"
	"github.com/bdware/tendermint/libs/log"
	"github.com/bdware/tendermint/libs/service"
	"github.com/bdware/tendermint/p2p"
	tmconn "github.com/bdware/tendermint/p2p/conn"
	"github.com/bdware/tendermint/p2p/libp2p/util"
)

var _ p2p.Peer = (*peer)(nil)

// peer implements p2p.Peer.
// Before using a peer, you will need to perform a handshake on connection.
type peer struct {
	service.BaseService

	outbound   bool
	persistent bool
	socketAddr *p2p.NetAddress

	// our local peer host to send msg to this peer
	host host.Host
	conn *tmconn.Libp2pMConnection

	// peer's node info and the channel it knows about
	// channels = nodeInfo.Channels
	// cached to avoid copying nodeInfo in hasChannel
	nodeInfo p2p.NodeInfo
	channels []byte

	// User data
	Data *cmap.CMap

	metrics       *p2p.Metrics
	metricsTicker *time.Ticker
}

type PeerOption func(*peer)

func newPeer(
	host host.Host,
	nodeInfo p2p.NodeInfo,
	reactorsByCh map[byte]p2p.Reactor,
	chDescs []*tmconn.ChannelDescriptor,
	onPeerError func(p2p.Peer, interface{}),
	options ...PeerOption,
) *peer {
	p := &peer{
		host:          host,
		nodeInfo:      nodeInfo,
		channels:      nodeInfo.(p2p.DefaultNodeInfo).Channels, // TODO
		Data:          cmap.NewCMap(),
		metricsTicker: time.NewTicker(p2p.MetricsTickerDuration),
		metrics:       p2p.NopMetrics(),
	}

	p.conn = createLpMConnection(
		host,
		p,
		reactorsByCh,
		chDescs,
		onPeerError,
	)
	p.BaseService = *service.NewBaseService(nil, "Peer", p)
	for _, option := range options {
		option(p)
	}

	return p
}

// String representation.
func (p *peer) String() string {
	if p.outbound {
		return fmt.Sprintf("Peer{%v %v out}", p.RemoteAddr(), p.ID())
	}

	return fmt.Sprintf("Peer{%v %v in}", p.RemoteAddr(), p.ID())
}

//---------------------------------------------------
// Implements service.Service

// SetLogger implements BaseService.
func (p *peer) SetLogger(l log.Logger) {
	p.Logger = l
	p.conn.SetLogger(l)
}

// OnStart implements BaseService.
func (p *peer) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}

	if err := p.conn.Start(); err != nil {
		return err
	}

	go p.metricsReporter()
	return nil
}

// FlushStop mimics OnStop but additionally ensures that all successful
// .Send() calls will get flushed before closing the connection.
// NOTE: it is not safe to call this method more than once.
func (p *peer) FlushStop() {
	p.metricsTicker.Stop()
	p.BaseService.OnStop()
	p.conn.FlushStop() // stop everything and close the conn
}

// OnStop implements BaseService.
func (p *peer) OnStop() {
	p.metricsTicker.Stop()
	p.BaseService.OnStop()
	p.conn.Stop() // stop everything and close the conn
}

//---------------------------------------------------
// Implements Peer

// ID returns the peer's ID
func (p *peer) ID() p2p.ID {
	return p.nodeInfo.ID()
}

func (p *peer) RemoteIP() net.IP {
	return p.socketAddr.IP
}

// IsOutbound returns true if the connection is outbound, false otherwise.
func (p *peer) IsOutbound() bool {
	return p.outbound
}

// IsPersistent returns true if the peer is persitent, false otherwise.
func (p *peer) IsPersistent() bool {
	return p.persistent
}

// NodeInfo returns a copy of the peer's NodeInfo.
func (p *peer) NodeInfo() p2p.NodeInfo {
	return p.nodeInfo
}

// SocketAddr returns the address of the socket.
// For outbound peers, it's the address dialed (after DNS resolution).
// For inbound peers, it's the address returned by the underlying connection
// (not what's reported in the peer's NodeInfo).
func (p *peer) SocketAddr() *p2p.NetAddress {
	return p.socketAddr
}

// Status returns the peer's ConnectionStatus. not used
func (p *peer) Status() tmconn.ConnectionStatus {
	return tmconn.ConnectionStatus{}
}

// Send msg bytes to the channel identified by chID byte. Returns false if the
// send queue is full after timeout, specified by MConnection.
func (p *peer) Send(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() {
		// see Switch#Broadcast, where we fetch the list of peers and loop over
		// them - while we're looping, one peer may be removed and stopped.
		return false
	} else if !p.hasChannel(chID) {
		return false
	}
	//s := p.stream
	//err := p.sendBytesTo(s, msgBytes, chID)
	res := p.conn.Send(chID, msgBytes)
	if res {
		labels := []string{
			"peer_id", string(p.ID()),
			"chID", fmt.Sprintf("%#x", chID),
		}
		p.metrics.PeerSendBytesTotal.With(labels...).Add(float64(len(msgBytes)))
	}
	return res
}

// TrySend msg bytes to the channel identified by chID byte. Immediately returns
// false if the send queue is full.
func (p *peer) TrySend(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() {
		return false
	} else if !p.hasChannel(chID) {
		return false
	}
	res := p.conn.TrySend(chID, msgBytes)
	if res {
		labels := []string{
			"peer_id", string(p.ID()),
			"chID", fmt.Sprintf("%#x", chID),
		}
		p.metrics.PeerSendBytesTotal.With(labels...).Add(float64(len(msgBytes)))
	}
	return res
}

// Get the data for a given key.
func (p *peer) Get(key string) interface{} {
	return p.Data.Get(key)
}

// Set sets the data for the given key.
func (p *peer) Set(key string, data interface{}) {
	p.Data.Set(key, data)
}

// hasChannel returns true if the peer reported
// knowing about the given chID.
func (p *peer) hasChannel(chID byte) bool {
	for _, ch := range p.channels {
		if ch == chID {
			return true
		}
	}
	// NOTE: probably will want to remove this
	// but could be helpful while the feature is new
	p.Logger.Debug(
		"Unknown channel for peer",
		"channel",
		chID,
		"channels",
		p.channels,
	)
	return false
}

// CloseConn closes libp2p connection
func (p *peer) CloseConn() error {
	return p.host.Network().ClosePeer(util.ID2Libp2pID(p.ID()))
}

//---------------------------------------------------
// methods only used for testing
// TODO: can we remove these?

// RemoteAddr returns peer's remote network address.
func (p *peer) RemoteAddr() net.Addr {
	// not sure
	return &net.TCPAddr{IP: p.socketAddr.IP, Port: (int)(p.socketAddr.Port)}
}

// CanSend returns true if the stream exists, false otherwise.
func (p *peer) CanSend(chID byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.conn.CanSend(chID)
}

//---------------------------------------------------

func PeerMetrics(metrics *p2p.Metrics) PeerOption {
	return func(p *peer) {
		p.metrics = metrics
	}
}

func (p *peer) metricsReporter() {
	for {
		select {
		case <-p.metricsTicker.C:
			// TODO: refactor this?
			//status := p.mconn.Status()
			//var sendQueueSize float64
			//for _, chStatus := range status.Channels {
			//	sendQueueSize += float64(chStatus.SendQueueSize)
			//}
			//
			//p.metrics.PeerPendingSendBytes.With("peer_id", string(p.ID())).Set(sendQueueSize)
		case <-p.Quit():
			return
		}
	}
}

//------------------------------------------------------------------
// helper funcs

func createConnection(
	s network.Stream,
	p *peer,
	reactorsByCh map[byte]p2p.Reactor,
	chDescs []*tmconn.ChannelDescriptor,
	onPeerError func(p2p.Peer, interface{}),
) *Connection {

	onReceive := func(chID byte, msgBytes []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			// Note that its ok to panic here as it's caught in the conn._recover,
			// which does onPeerError.
			panic(fmt.Sprintf("Unknown channel %X", chID))
		}
		labels := []string{
			"peer_id", string(p.ID()),
			"chID", fmt.Sprintf("%#x", chID),
		}
		p.metrics.PeerReceiveBytesTotal.With(labels...).Add(float64(len(msgBytes)))
		reactor.Receive(chID, p, msgBytes)
	}

	onError := func(r interface{}) {
		onPeerError(p, r)
	}

	return NewConnection(
		s,
		chDescs,
		onReceive,
		onError,
	)
}

func createLpMConnection(
	host host.Host,
	p *peer,
	reactorsByCh map[byte]p2p.Reactor,
	chDescs []*tmconn.ChannelDescriptor,
	onPeerError func(p2p.Peer, interface{}),
) *tmconn.Libp2pMConnection{

	onReceive := func(chID byte, msgBytes []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			// Note that its ok to panic here as it's caught in the conn._recover,
			// which does onPeerError.
			panic(fmt.Sprintf("Unknown channel %X", chID))
		}
		labels := []string{
			"peer_id", string(p.ID()),
			"chID", fmt.Sprintf("%#x", chID),
		}
		p.metrics.PeerReceiveBytesTotal.With(labels...).Add(float64(len(msgBytes)))
		reactor.Receive(chID, p, msgBytes)
	}

	onError := func(r interface{}) {
		onPeerError(p, r)
	}

	return tmconn.NewLibp2pMConnection(
		"",
		host,
		util.ID2Libp2pID(p.ID()),
		chDescs,
		onReceive,
		onError,
	)
}
