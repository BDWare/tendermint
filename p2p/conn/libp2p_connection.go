package conn

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/tendermint/go-amino"

	flow "github.com/bdware/tendermint/libs/flowrate"
	tmmath "github.com/bdware/tendermint/libs/math"
	"github.com/bdware/tendermint/libs/service"
	"github.com/bdware/tendermint/libs/timer"
)

type recvMsg struct {
	chID     byte
	msgBytes []byte
}

type Libp2pMConnection struct {
	*MConnection

	protocolPrefix string
	host           host.Host
	peerID         peer.ID
	channels       []*Libp2pChannel
	channelsIdx    map[byte]*Libp2pChannel
	// Stream for the ping protocol.
	stream network.Stream
	// One stream, bufio.Reader and bufio.Writer for each stream.
	streams        map[byte]network.Stream
	bufConnReaders map[byte]*bufio.Reader
	bufConnWriters map[byte]*bufio.Writer
	// recv channel to serialize received messages on all streams to call onReceive().
	recv chan recvMsg
}

// NewLibp2pMConnection wraps net.Conn and creates multiplex connection
func NewLibp2pMConnection(
	protocolPrefix string,
	host host.Host,
	peerID peer.ID,
	pingStream network.Stream,
	chStreams map[byte]network.Stream,
	chDescs []*ChannelDescriptor,
	onReceive receiveCbFunc,
	onError errorCbFunc,
) *Libp2pMConnection {
	return NewLibp2pMConnectionWithConfig(
		protocolPrefix,
		host,
		peerID,
		pingStream,
		chStreams,
		chDescs,
		onReceive,
		onError,
		DefaultMConnConfig(),
	)

}

// NewLibp2pMConnectionWithConfig wraps net.Conn and creates multiplex connection with a config
func NewLibp2pMConnectionWithConfig(
	protocolPrefix string,
	host host.Host,
	peerID peer.ID,
	pingStream network.Stream,
	chStreams map[byte]network.Stream,
	chDescs []*ChannelDescriptor,
	onReceive receiveCbFunc,
	onError errorCbFunc,
	config MConnConfig,
) *Libp2pMConnection {
	if config.PongTimeout >= config.PingInterval {
		panic("pongTimeout must be less than pingInterval (otherwise, next ping will reset pong timer)")
	}

	mconn := &Libp2pMConnection{
		MConnection: &MConnection{
			sendMonitor: flow.New(0, 0),
			recvMonitor: flow.New(0, 0),
			send:        make(chan struct{}, 1),
			pong:        make(chan struct{}, 1),
			onReceive:   onReceive,
			onError:     onError,
			config:      config,
			created:     time.Now(),
		},
		protocolPrefix: protocolPrefix,
		host:           host,
		peerID:         peerID,
		bufConnReaders: map[byte]*bufio.Reader{},
		bufConnWriters: map[byte]*bufio.Writer{},
		stream:         pingStream,
		streams:        chStreams,
	}

	// Create channels
	var lpChannelsIdx = map[byte]*Libp2pChannel{}
	var ChannelsIdx = map[byte]*Channel{}
	var lpChannels = []*Libp2pChannel{}
	var Channels = []*Channel{}

	for _, desc := range chDescs {
		id := desc.ID
		lpChannel := newLibp2pChannel(mconn, *desc)
		lpChannelsIdx[id] = lpChannel
		ChannelsIdx[id] = lpChannel.Channel
		lpChannels = append(lpChannels, lpChannel)
		Channels = append(Channels, lpChannel.Channel)
	}
	mconn.channels = lpChannels
	mconn.channelsIdx = lpChannelsIdx
	mconn.MConnection.channels = Channels
	mconn.MConnection.channelsIdx = ChannelsIdx

	mconn.BaseService = *service.NewBaseService(nil, "Libp2pMConnection", mconn)

	// maxPacketMsgSize() is a bit heavy, so call just once
	mconn._maxPacketMsgSize = mconn.maxPacketMsgSize()

	return mconn
}

//func (c *Libp2pMConnection) SetLogger(l log.Logger) {
//	c.MConnection.SetLogger(l)
//}

// OnStart implements BaseService
func (c *Libp2pMConnection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}
	// 1 min timeout for all streams.
	//ctx, cncl := context.WithTimeout(context.Background(), 1*time.Minute)
	//go func() {
	//	<-c.Quit()
	//	cncl()
	//}()

	// Create stream, bufio.Reader and bufio.Writer for the ping protocol.
	//s, err := c.host.NewStream(ctx, c.peerID, protocol.ID(c.protocolPrefix+PingProtocolID))
	//if err != nil {
	//	return fmt.Errorf("failed to call host.NewStream to peer %v: %v", c.peerID, err)
	//}
	//c.stream = s
	s := c.stream
	c.bufConnReader = bufio.NewReaderSize(s, minReadBufferSize)
	c.bufConnWriter = bufio.NewWriterSize(s, minWriteBufferSize)

	// Create stream, bufio.Reader and bufio.Writer for each channel.
	//for _, ch := range c.channels {
	//	chID := ch.desc.ID
	//	s, err := c.host.NewStream(ctx, c.peerID, protocol.ID(c.protocolPrefix+string(chID)))
	//	if err != nil {
	//		return fmt.Errorf("failed to call host.NewStream to peer %v: %v", c.peerID, err)
	//	}
	//	c.streams[chID] = s
	//	c.bufConnReaders[chID] = bufio.NewReaderSize(s, minReadBufferSize)
	//	c.bufConnWriters[chID] = bufio.NewWriterSize(s, minWriteBufferSize)
	//}

	for chID, s := range c.streams {
		c.bufConnReaders[chID] = bufio.NewReaderSize(s, minReadBufferSize)
		c.bufConnWriters[chID] = bufio.NewWriterSize(s, minWriteBufferSize)
	}

	c.recv = make(chan recvMsg)

	// TODO: Maybe we can not run ping protocol in Tendermint
	// and run it in libp2p outside of Tendermint if we need.
	c.flushTimer = timer.NewThrottleTimer("flush", c.config.FlushThrottle)
	c.pingTimer = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTimer = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})

	// Start all the routines.
	go c.sendRoutine()
	go c.recvRoutinePingpong()
	go c.recvRoutine()
	for _, ch := range c.channels {
		go c.recvRoutineChannel(ch)
	}

	return nil
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
// It uses the stopMtx to ensure only one of FlushStop and OnStop can do this at a time.
//func (c *Libp2pMConnection) stopServices() (alreadyStopped bool) {
//	return c.MConnection.stopServices()
//}

// FlushStop replicates the logic of OnStop.
// It additionally ensures that all successful
// .Send() calls will get flushed before closing
// the connection.
func (c *Libp2pMConnection) FlushStop() {
	if c.stopServices() {
		return
	}

	// this block is unique to FlushStop
	{
		// wait until the sendRoutine exits
		// so we dont race on calling sendSomePacketMsgs
		<-c.doneSendRoutine

		// Send and flush all pending msgs.
		// Since sendRoutine has exited, we can call this
		// safely
		eof := c.sendSomePacketMsgs()
		for !eof {
			eof = c.sendSomePacketMsgs()
		}
		c.flush()

		// Now we can close the connection
	}

	// Streams already closed in recvRoutineX routines.
	//c.stream.Reset()
	//for _, s := range c.streams {
	//	s.Reset()
	//}

	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.

	// c.Stop()
}

// OnStop implements BaseService
func (c *Libp2pMConnection) OnStop() {
	if c.stopServices() {
		return
	}

	// Streams already closed in recvRoutineX routines.
	//c.stream.Reset()
	//for _, s := range c.streams {
	//	s.Reset()
	//}

	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.
}

func (c *Libp2pMConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.peerID)
}

func (c *Libp2pMConnection) flush() {
	// Flush ping stream.
	c.MConnection.flush()
	// Flush channel streams.
	for _, w := range c.bufConnWriters {
		if err := w.Flush(); err != nil {
			c.Logger.Error("Libp2pMConnection flush failed for a channel", "err", err)
		}
	}
}

// Catch panics, usually caused by remote disconnects.
//func (c *Libp2pMConnection) _recover() {
//	c.MConnection._recover()
//}

//func (c *Libp2pMConnection) stopForError(r interface{}) {
//	c.MConnection.stopForError(r)
//}

// Queues a message to be sent to channel.
//func (c *Libp2pMConnection) Send(chID byte, msgBytes []byte) bool {
//	return c.MConnection.Send(chID, msgBytes)
//}

// Queues a message to be sent to channel.
// Nonblocking, returns true if successful.
//func (c *Libp2pMConnection) TrySend(chID byte, msgBytes []byte) bool {
//	return c.MConnection.TrySend(chID, msgBytes)
//}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
//func (c *Libp2pMConnection) CanSend(chID byte) bool {
//	return c.MConnection.CanSend(chID)
//}

// sendRoutine polls for packets to send from channels.
//
// #BDWare: This function is not modified yed,
// it needs to be copied because it calls flush() and indirectly sendPacketMsg() which are modified.
func (c *Libp2pMConnection) sendRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		var _n int64
		var err error
	SELECTION:
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufConnWriter.
			// Flush all streams here.
			c.flush()
		case <-c.chStatsTimer.C:
			for _, channel := range c.channels {
				channel.updateStats()
			}
		case <-c.pingTimer.C:
			c.Logger.Debug("Send Ping")
			_n, err = cdc.MarshalBinaryLengthPrefixedWriter(c.bufConnWriter, PacketPing{})
			if err != nil {
				break SELECTION
			}
			c.sendMonitor.Update(int(_n))
			c.Logger.Debug("Starting pong timer", "dur", c.config.PongTimeout)
			c.pongTimer = time.AfterFunc(c.config.PongTimeout, func() {
				select {
				case c.pongTimeoutCh <- true:
				default:
				}
			})
			// Only flush ping stream here.
			c.MConnection.flush()
		case timeout := <-c.pongTimeoutCh:
			if timeout {
				c.Logger.Debug("Pong timeout")
				err = errors.New("pong timeout")
			} else {
				c.stopPongTimer()
			}
		case <-c.pong:
			c.Logger.Debug("Send Pong")
			_n, err = cdc.MarshalBinaryLengthPrefixedWriter(c.bufConnWriter, PacketPong{})
			if err != nil {
				break SELECTION
			}
			c.sendMonitor.Update(int(_n))
			// Only flush ping stream here.
			c.MConnection.flush()
		case <-c.quitSendRoutine:
			break FOR_LOOP
		case <-c.send:
			// Send some PacketMsgs
			eof := c.sendSomePacketMsgs()
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			c.Logger.Error("Connection failed @ sendRoutine", "conn", c, "err", err)
			c.stopForError(err)
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
//
// #BDWare: This function is not modified yed,
// it needs to be copied because it calls sendPacketMsg() which is modified.
func (c *Libp2pMConnection) sendSomePacketMsgs() bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.SendRate), true)

	// Now send some PacketLibp2pMsgs.
	for i := 0; i < numBatchPacketMsgs; i++ {
		if c.sendPacketMsg() {
			return true
		}
	}
	return false
}

// Returns true if messages from channels were exhausted.
func (c *Libp2pMConnection) sendPacketMsg() bool {
	// Choose a channel to create a PacketLibp2pMsg from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *Libp2pChannel
	for _, channel := range c.channels {
		// If nothing to send, skip this channel
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		ratio := float32(channel.recentlySent) / float32(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}

	// Nothing to send?
	if leastChannel == nil {
		return true
	}
	// c.Logger.Info("Found a msgPacket to send")

	// Make & send a PacketLibp2pMsg from this channel
	w, ok := c.bufConnWriters[leastChannel.desc.ID]
	if !ok {
		err := fmt.Errorf("stream not found for channel %v", leastChannel.desc.ID)
		c.Logger.Error("Failed to write PacketLibp2pMsg", "err", err)
		c.stopForError(err)
		return true
	}
	_n, err := leastChannel.writePacketMsgTo(w)
	if err != nil {
		c.Logger.Error("Failed to write PacketLibp2pMsg", "err", err)
		c.stopForError(err)
		return true
	}
	c.sendMonitor.Update(int(_n))
	c.flushTimer.Set()
	return false
}

// recvRoutinePingpong reads ping-pong messages.
// Blocks depending on how the connection is throttled.
// Otherwise, it never blocks.
func (c *Libp2pMConnection) recvRoutinePingpong() {
	defer c._recover()

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(tmmath.MinInt(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet Packet
		var _n int64
		var err error
		_n, err = cdc.UnmarshalBinaryLengthPrefixedReader(c.bufConnReader, &packet, int64(c._maxPacketMsgSize))
		c.recvMonitor.Update(int(_n))

		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is excpected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if err == io.EOF {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)", "conn", c)
				} else {
					c.Logger.Error("Connection failed @ recvRoutine (reading byte)", "conn", c, "err", err)
				}
				c.stopForError(err)
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch packet.(type) {
		case PacketPing:
			// TODO: prevent abuse, as they cause flush()'s.
			// https://github.com/tendermint/tendermint/issues/1190
			c.Logger.Debug("Receive Ping")
			select {
			case c.pong <- struct{}{}:
			default:
				// never block
			}
		case PacketPong:
			c.Logger.Debug("Receive Pong")
			select {
			case c.pongTimeoutCh <- false:
			default:
				// never block
			}
		default:
			err := fmt.Errorf("unexpected message type %v", reflect.TypeOf(packet))
			c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
			c.stopForError(err)
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stream.Reset()
	close(c.pong)
	for range c.pong {
		// Drain
	}
}

// recvRoutine receives the message from c.recv and serially calls onReceive().
func (c *Libp2pMConnection) recvRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		select {
		case <-c.quitRecvRoutine:
			break FOR_LOOP
		case m := <-c.recv:
			// NOTE: This means the reactor.Receive runs in the same thread as the p2p recv routine
			c.onReceive(m.chID, m.msgBytes)
		}
	}

	// Cleanup
	close(c.recv)
	for range c.recv {
		// Drain
	}
}

// recvRoutineChannel reads PacketLibp2pMsgs and reconstructs the message using the channels' "recving" buffer.
// After a whole message has been assembled, it's sent to c.recv to serially call onReceive().
// Blocks depending on how the connection is throttled.
// Otherwise, it never blocks.
func (c *Libp2pMConnection) recvRoutineChannel(ch *Libp2pChannel) {
	defer c._recover()

	chID := ch.desc.ID
	// Get the bufio.Reader for the channel.
	s, sok := c.streams[chID]
	r, rok := c.bufConnReaders[chID]
	if !sok || !rok {
		err := fmt.Errorf("stream not found for channel %v", chID)
		c.Logger.Error("Failed to start recvRoutineChannel", "err", err)
		c.stopForError(err)
		panic(err)
	}

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(tmmath.MinInt(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet Packet
		var _n int64
		var err error
		_n, err = cdc.UnmarshalBinaryLengthPrefixedReader(r, &packet, int64(c._maxPacketMsgSize))
		c.recvMonitor.Update(int(_n))

		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is excpected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if err == io.EOF {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)", "conn", c)
				} else {
					c.Logger.Error("Connection failed @ recvRoutine (reading byte)", "conn", c, "err", err)
				}
				c.stopForError(err)
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pkt := packet.(type) {
		case PacketLibp2pMsg:
			msgBytes, err := ch.recvPacketMsg(pkt)
			if err != nil {
				if c.IsRunning() {
					c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
					c.stopForError(err)
				}
				break FOR_LOOP
			}
			if msgBytes != nil {
				c.Logger.Debug("Received bytes", "chID", chID, "msgBytes", fmt.Sprintf("%X", msgBytes))
				// Send to c.recv to serialize received messages on all streams to call onReceive().
				select {
				case <-c.quitRecvRoutine:
					break FOR_LOOP
				case c.recv <- recvMsg{chID, msgBytes}:
				}
			}
		default:
			err := fmt.Errorf("unexpected message type %v", reflect.TypeOf(packet))
			c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
			c.stopForError(err)
			break FOR_LOOP
		}
	}

	// Cleanup
	s.Reset()
}

// not goroutine-safe
//func (c *Libp2pMConnection) stopPongTimer() {
//	c.MConnection.stopPongTimer()
//}

// maxPacketMsgSize returns a maximum size of PacketMsg, including the overhead
// of amino encoding.
func (c *Libp2pMConnection) maxPacketMsgSize() int {
	return len(cdc.MustMarshalBinaryLengthPrefixed(PacketLibp2pMsg{
		EOF:   1,
		Bytes: make([]byte, c.config.MaxPacketMsgPayloadSize),
	})) + 10 // leave room for changes in amino
}

//func (c *Libp2pMConnection) Status() ConnectionStatus {
//	return c.MConnection.Status()
//}

//-----------------------------------------------------------------------------

// TODO: lowercase.
// NOTE: not goroutine-safe.
type Libp2pChannel struct {
	*Channel
	conn *Libp2pMConnection
}

func newLibp2pChannel(conn *Libp2pMConnection, desc ChannelDescriptor) *Libp2pChannel {
	return &Libp2pChannel{
		Channel: newChannel(conn.MConnection, desc),
		conn:    conn,
	}
}

//func (ch *Libp2pChannel) SetLogger(l log.Logger) {
//	ch.Channel.SetLogger(l)
//}

// Queues message to send to this channel.
// Goroutine-safe
// Times out (and returns false) after defaultSendTimeout
//func (ch *Libp2pChannel) sendBytes(bytes []byte) bool {
//	return ch.Channel.sendBytes(bytes)
//}

// Queues message to send to this channel.
// Nonblocking, returns true if successful.
// Goroutine-safe
//func (ch *Libp2pChannel) trySendBytes(bytes []byte) bool {
//	return ch.Channel.trySendBytes(bytes)
//}

// Goroutine-safe
//func (ch *Libp2pChannel) loadSendQueueSize() (size int) {
//	return ch.Channel.loadSendQueueSize()
//}

// Goroutine-safe
// Use only as a heuristic.
//func (ch *Libp2pChannel) canSend() bool {
//	return ch.Channel.canSend()
//}

// Returns true if any PacketLibp2pMsg are pending to be sent.
// Call before calling nextPacketMsg()
// Goroutine-safe
//func (ch *Libp2pChannel) isSendPending() bool {
//	return ch.Channel.isSendPending()
//}

// Creates a new PacketLibp2pMsg to send.
// Not goroutine-safe
func (ch *Libp2pChannel) nextPacketMsg() PacketLibp2pMsg {
	packet := PacketLibp2pMsg{}
	maxSize := ch.maxPacketMsgPayloadSize
	packet.Bytes = ch.sending[:tmmath.MinInt(maxSize, len(ch.sending))]
	if len(ch.sending) <= maxSize {
		packet.EOF = byte(0x01)
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		packet.EOF = byte(0x00)
		ch.sending = ch.sending[tmmath.MinInt(maxSize, len(ch.sending)):]
	}
	return packet
}

// Writes next PacketLibp2pMsg to w and updates c.recentlySent.
// Not goroutine-safe
//
// #BDWare: This function is not modified yed,
// it needs to be copied because it calls nextPacketMsg() which is modified.
func (ch *Libp2pChannel) writePacketMsgTo(w io.Writer) (n int64, err error) {
	var packet = ch.nextPacketMsg()
	n, err = cdc.MarshalBinaryLengthPrefixedWriter(w, packet)
	atomic.AddInt64(&ch.recentlySent, n)
	return
}

// Handles incoming PacketLibp2pMsgs. It returns a message bytes if message is
// complete. NOTE message bytes may change on next call to recvPacketMsg.
// Not goroutine-safe
func (ch *Libp2pChannel) recvPacketMsg(packet PacketLibp2pMsg) ([]byte, error) {
	ch.Logger.Debug("Read PacketLibp2pMsg", "conn", ch.conn, "packet", packet)
	var recvCap, recvReceived = ch.desc.RecvMessageCapacity, len(ch.recving) + len(packet.Bytes)
	if recvCap < recvReceived {
		return nil, fmt.Errorf("received message exceeds available capacity: %v < %v", recvCap, recvReceived)
	}
	ch.recving = append(ch.recving, packet.Bytes...)
	if packet.EOF == byte(0x01) {
		msgBytes := ch.recving

		// clear the slice without re-allocating.
		// http://stackoverflow.com/questions/16971741/how-do-you-clear-a-slice-in-go
		//   suggests this could be a memory leak, but we might as well keep the memory for the channel until it closes,
		//	at which point the recving slice stops being used and should be garbage collected
		ch.recving = ch.recving[:0] // make([]byte, 0, ch.desc.RecvBufferCapacity)
		return msgBytes, nil
	}
	return nil, nil
}

// Call this periodically to update stats for throttling purposes.
// Not goroutine-safe
//func (ch *Libp2pChannel) updateStats() {
//	ch.Channel.updateStats()
//}

//----------------------------------------
// Packet

func RegisterLibp2pPacket(cdc *amino.Codec) {
	cdc.RegisterConcrete(PacketLibp2pMsg{}, "tendermint/p2p/PacketLibp2pMsg", nil)
}

func (PacketLibp2pMsg) AssertIsPacket() {}

type PacketLibp2pMsg struct {
	//ChannelID byte // Not needed for libp2p
	EOF   byte // 1 means message ends here.
	Bytes []byte
}

func (mp PacketLibp2pMsg) String() string {
	return fmt.Sprintf("PacketLibp2pMsg{%X T:%X}", mp.Bytes, mp.EOF)
}
