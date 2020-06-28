package p2p

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"net"
	"time"

	"github.com/bdware/tendermint/libs/cmap"
	"github.com/bdware/tendermint/libs/log"
	"github.com/bdware/tendermint/libs/service"

	tmconn "github.com/bdware/tendermint/p2p/conn"
)

var _ Peer = (*lpPeer)(nil)

// lpPeer implements Peer.
//
// Before using a lpPeer, you will need to perform a handshake on connection.
type lpPeer struct {
	service.BaseService

	outbound   bool
	persistent bool
	socketAddr *NetAddress

	// lpPeer's node info and the channel it knows about
	// channels = nodeInfo.Channels
	// cached to avoid copying nodeInfo in hasChannel
	nodeInfo NodeInfo
	channels []byte

	streams map[byte]network.Stream
	// our local peer host to send msg to this lpPeer
	host    host.Host

	// User data
	Data *cmap.CMap

	metrics       *Metrics
	metricsTicker *time.Ticker

	onReceive func(chID byte, msgBytes []byte)
}

func (p *lpPeer) RemoteIP() net.IP {
	return p.socketAddr.IP
}

func newLpPeer(
	nodeInfo NodeInfo,
	reactorsByCh map[byte]Reactor,
	host host.Host,
	chDescs []*tmconn.ChannelDescriptor,
	onPeerError func(Peer, interface{}),
) *lpPeer {
	p := &lpPeer{
		nodeInfo:      nodeInfo,
		channels:      nodeInfo.(DefaultNodeInfo).Channels, // TODO
		streams:       make(map[byte]network.Stream),
		host:          host,
		Data:          cmap.NewCMap(),
		metricsTicker: time.NewTicker(metricsTickerDuration),
		metrics:       NopMetrics(),
	}

	p.onReceive = func(chID byte, msgBytes []byte) {
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
	//p.mconn = createMConnection(
	//	p,
	//	reactorsByCh,
	//	chDescs,
	//	onPeerError,
	//)
	p.BaseService = *service.NewBaseService(nil, "Peer", p)
	//for _, option := range options {
	//	option(p)
	//}

	return p
}

// String representation.
func (p *lpPeer) String() string {
	if p.outbound {
		return fmt.Sprintf("Peer{%v %v out}", p.RemoteAddr(), p.ID())
	}

	return fmt.Sprintf("Peer{%v %v in}", p.RemoteAddr(), p.ID())
}

func (p *lpPeer) getStream(chID byte) network.Stream {
	s, ok := p.streams[chID]
	if ok {
		return s
	} else {
		s, err := p.host.NewStream(context.TODO(), ID2lpID(p.nodeInfo.ID()), protocolForChannel(chID))
		if err != nil {
			p.Logger.Info("fail to create a stream for channel %d", chID)
			return nil
		}
		p.streams[chID] = s
		go p.recvRoutine(s, chID)
		return s
	}
}

//---------------------------------------------------
// Implements service.Service

// SetLogger implements BaseService.
func (p *lpPeer) SetLogger(l log.Logger) {
	p.Logger = l
}

// OnStart implements BaseService.
func (p *lpPeer) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}

	go p.metricsReporter()
	return nil
}

// FlushStop mimics OnStop but additionally ensures that all successful
// .Send() calls will get flushed before closing the connection.
// NOTE: it is not safe to call this method more than once.
func (p *lpPeer) FlushStop() {
	p.metricsTicker.Stop()
	p.BaseService.OnStop()
}

// OnStop implements BaseService.
func (p *lpPeer) OnStop() {
	p.metricsTicker.Stop()
	p.BaseService.OnStop()
}

//---------------------------------------------------
// Implements Peer

// ID returns the lpPeer's ID
func (p *lpPeer) ID() ID {
	return p.nodeInfo.ID()
}

// IsOutbound returns true if the connection is outbound, false otherwise.
func (p *lpPeer) IsOutbound() bool {
	return p.outbound
}

// IsPersistent returns true if the lpPeer is persitent, false otherwise.
func (p *lpPeer) IsPersistent() bool {
	return p.persistent
}

// NodeInfo returns a copy of the lpPeer's NodeInfo.
func (p *lpPeer) NodeInfo() NodeInfo {
	return p.nodeInfo
}

// SocketAddr returns the address of the socket.
// For outbound peers, it's the address dialed (after DNS resolution).
// For inbound peers, it's the address returned by the underlying connection
// (not what's reported in the lpPeer's NodeInfo).
func (p *lpPeer) SocketAddr() *NetAddress {
	return p.socketAddr
}

// Status returns the lpPeer's ConnectionStatus. not used
func (p *lpPeer) Status() tmconn.ConnectionStatus {
	return tmconn.ConnectionStatus{}
}

// Send msg bytes to the channel identified by chID byte.
func (p *lpPeer) Send(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() {
		// see LpSwitch#Broadcast, where we fetch the list of peers and loop over
		// them - while we're looping, one lpPeer may be removed and stopped.
		return false
	} else if !p.hasChannel(chID) {
		return false
	}
	//res := p.mconn.Send(chID, msgBytes)
	s := p.getStream(chID)
	if s == nil {
		return false
	}
	err := p.sendBytesTo(s, msgBytes, chID)
	if err == nil {
		labels := []string{
			"peer_id", string(p.ID()),
			"chID", fmt.Sprintf("%#x", chID),
		}
		p.metrics.PeerSendBytesTotal.With(labels...).Add(float64(len(msgBytes)))
	}
	return err == nil
}

// TrySend is the same as Send for lpPeer
func (p *lpPeer) TrySend(chID byte, msgBytes []byte) bool {
	return p.Send(chID, msgBytes)
}

func readUvarint(r io.Reader) (uint64, error) {
	var x uint64
	var s uint
	for i := 0; ; i++ {
		buf := make([]byte, 1)
		n, err := r.Read(buf)
		if err != nil || n != 1 {
			return x, err
		}
		b := buf[0]
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return x, fmt.Errorf("overflow")
			}
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}
func (p *lpPeer) recvRoutine(s network.Stream, chID byte) {
	for {
		// binary.ReadUvarint
		length, err := readUvarint(s)
		if err != nil {
			break
		}

		buf := make([]byte, length)
		_, err = io.ReadFull(s, buf)
		if err != nil {
			break
		}
		p.onReceive(chID, buf)
	}
}

// Get the data for a given key.
func (p *lpPeer) Get(key string) interface{} {
	return p.Data.Get(key)
}

// Set sets the data for the given key.
func (p *lpPeer) Set(key string, data interface{}) {
	p.Data.Set(key, data)
}

// hasChannel returns true if the lpPeer reported
// knowing about the given chID.
func (p *lpPeer) hasChannel(chID byte) bool {
	for _, ch := range p.channels {
		if ch == chID {
			return true
		}
	}
	// NOTE: probably will want to remove this
	// but could be helpful while the feature is new
	p.Logger.Debug(
		"Unknown channel for lpPeer",
		"channel",
		chID,
		"channels",
		p.channels,
	)
	return false
}

// CloseConn closes libp2p connection
func (p *lpPeer) CloseConn() error {
	// do nothing now
	return p.host.Network().ClosePeer(ID2lpID(p.ID()))
}

func writeUvarint(w io.Writer, i uint64) error {
	varintbuf := make([]byte, 16)
	n := binary.PutUvarint(varintbuf, i)
	_, err := w.Write(varintbuf[:n])
	if err != nil {
		return err
	}
	return nil
}

func (p *lpPeer) sendBytesTo(s network.Stream, msg []byte, chID byte) error {
	ln := uint64(len(msg))
	err := writeUvarint(s, ln)
	if err == nil {
		_, err = s.Write(msg)
	}
	return err
}

//---------------------------------------------------
// methods only used for testing
// TODO: can we remove these?

// RemoteAddr returns lpPeer's remote network address.
func (p *lpPeer) RemoteAddr() net.Addr {
	// not sure
	return &net.TCPAddr{IP: p.socketAddr.IP, Port: (int)(p.socketAddr.Port)}

}

// CanSend returns true if the stream exists, false otherwise.
func (p *lpPeer) CanSend(chID byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.getStream(chID) != nil
}

//---------------------------------------------------

//func PeerMetrics(metrics *Metrics) PeerOption {
//	return func(p *lpPeer) {
//		p.metrics = metrics
//	}
//}

func (p *lpPeer) metricsReporter() {
	for {
		select {
		case <-p.metricsTicker.C:
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

//func createMConnection(
//	conn net.Conn,
//	p *lpPeer,
//	reactorsByCh map[byte]Reactor,
//	chDescs []*tmconn.ChannelDescriptor,
//	onPeerError func(Peer, interface{}),
//	config tmconn.MConnConfig,
//) *tmconn.MConnection {
//
//	onReceive := func(chID byte, msgBytes []byte) {
//		reactor := reactorsByCh[chID]
//		if reactor == nil {
//			// Note that its ok to panic here as it's caught in the conn._recover,
//			// which does onPeerError.
//			panic(fmt.Sprintf("Unknown channel %X", chID))
//		}
//		labels := []string{
//			"peer_id", string(p.ID()),
//			"chID", fmt.Sprintf("%#x", chID),
//		}
//		p.metrics.PeerReceiveBytesTotal.With(labels...).Add(float64(len(msgBytes)))
//		reactor.Receive(chID, p, msgBytes)
//	}
//
//	onError := func(r interface{}) {
//		onPeerError(p, r)
//	}
//
//	return tmconn.NewMConnectionWithConfig(
//		conn,
//		chDescs,
//		onReceive,
//		onError,
//		config,
//	)
//}
