package libp2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/pkg/errors"

	"github.com/bdware/tendermint/libs/log"
	"github.com/bdware/tendermint/libs/service"
	mconn "github.com/bdware/tendermint/p2p/conn"
)

type receiveCbFunc func(chID byte, msgBytes []byte)
type errorCbFunc func(interface{})

type Connection struct {
	service.BaseService

	s network.Stream

	// TODO: not used in libp2p yet (one send queue per channel with high concurrency)
	//channels      []*Channel
	//channelsIdx   map[byte]*Channel
	onReceive receiveCbFunc
	onError   errorCbFunc
	errored   uint32
	// TODO: not used in libp2p yet (not sure if need it)
	//config        mconn.MConnConfig

	// Closing quitSendRoutine will cause the sendRoutine to eventually quit.
	// doneSendRoutine is closed when the sendRoutine actually quits.
	//quitSendRoutine chan struct{}
	//doneSendRoutine chan struct{}

	// Closing quitRecvRouting will cause the recvRouting to eventually quit.
	quitRecvRoutine chan struct{}

	// used to ensure FlushStop and OnStop
	// are safe to call concurrently.
	stopMtx sync.Mutex

	created time.Time // time of creation

	// TODO: play ping pong (using *ping.PingService or ?)
}

// NewConnection wraps network.Stream and creates connection
func NewConnection(
	s network.Stream,
	chDescs []*mconn.ChannelDescriptor,
	onReceive receiveCbFunc,
	onError errorCbFunc,
) *Connection {
	return NewConnectionWithConfig(
		s,
		chDescs,
		onReceive,
		onError,
		mconn.DefaultMConnConfig(),
	)
}

func NewConnectionWithConfig(s network.Stream,
	chDescs []*mconn.ChannelDescriptor,
	onReceive receiveCbFunc,
	onError errorCbFunc,
	config mconn.MConnConfig,
) *Connection {
	conn := &Connection{
		s:         s,
		onReceive: onReceive,
		onError:   onError,
		created:   time.Now(),
	}

	conn.BaseService = *service.NewBaseService(nil, "Connection", conn)

	return conn
}

func (c *Connection) SetLogger(l log.Logger) {
	c.BaseService.SetLogger(l)
}

// OnStart implements BaseService
func (c *Connection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}

	c.quitRecvRoutine = make(chan struct{})
	go c.recvRoutine()
	return nil
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
// It uses the stopMtx to ensure only one of FlushStop and OnStop can do this at a time.
func (c *Connection) stopServices() (alreadyStopped bool) {
	c.stopMtx.Lock()
	defer c.stopMtx.Unlock()

	select {
	case <-c.quitRecvRoutine:
		// already quit
		return true
	default:
	}

	c.BaseService.OnStop()

	// inform the recvRouting that we are shutting down
	close(c.quitRecvRoutine)
	return false
}

// FlushStop replicates the logic of OnStop.
// It additionally ensures that all successful
// .Send() calls will get flushed before closing
// the connection.
func (c *Connection) FlushStop() {
	if c.stopServices() {
		return
	}

	c.s.Close() // nolint: errcheck

	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.

	// c.Stop()
}

// OnStop implements BaseService
func (c *Connection) OnStop() {
	if c.stopServices() {
		return
	}

	c.s.Close() // nolint: errcheck

	// We can't close pong safely here because
	// recvRoutine may write to it after we've stopped.
	// Though it doesn't need to get closed at all,
	// we close it @ recvRoutine.
}

func (c *Connection) String() string {
	return fmt.Sprintf("MConn{%v}", c.s.Conn().RemoteMultiaddr())
}

// Catch panics, usually caused by remote disconnects.
func (c *Connection) _recover() {
	if r := recover(); r != nil {
		c.Logger.Error("MConnection panicked", "err", r, "stack", string(debug.Stack()))
		c.stopForError(errors.Errorf("recovered from panic: %v", r))
	}
}

func (c *Connection) stopForError(r interface{}) {
	c.Stop()
	if atomic.CompareAndSwapUint32(&c.errored, 0, 1) {
		if c.onError != nil {
			c.onError(r)
		}
	}
}

// Queues a message to be sent to channel.
func (c *Connection) Send(chID byte, msgBytes []byte) bool {
	if !c.IsRunning() {
		return false
	}

	c.Logger.Debug("Send", "channel", chID, "conn", c, "msgBytes", fmt.Sprintf("%X", msgBytes))

	err := c.sendBytes(chID, msgBytes)
	// Send message to channel.
	if err == nil {
		return true
	} else {
		c.Logger.Debug("Send failed", "channel", chID, "conn", c, "msgBytes", fmt.Sprintf("%X", msgBytes))
		return false
	}
}

// Queues a message to be sent to channel.
// Nonblocking, returns true if successful.
// Because we don't implement queues of conn, TrySend is same as Send.
func (c *Connection) TrySend(chID byte, msgBytes []byte) bool {
	return c.Send(chID, msgBytes)
}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
func (c *Connection) CanSend(chID byte) bool {
	// Currently always true
	return true
}

func (c *Connection) recvRoutine() {
	defer c._recover()

FOR_LOOP:
	for {
		// binary.ReadUvarint
		length, err := readUvarint(c.s)
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

		buf := make([]byte, length)
		n, err := io.ReadFull(c.s, buf)
		if err != nil {
			if c.IsRunning() {
				c.Logger.Error("Connection failed @ recvRoutine", "conn", c, "err", err)
				c.stopForError(err)
			}
			break FOR_LOOP
		}
		chID := buf[0]
		if n != 0 {
			c.Logger.Debug("Received bytes", "chID", chID, "msgBytes", fmt.Sprintf("%X", buf[1:]))
			// Just leave checking channel ID to Receive
			c.onReceive(chID, buf[1:])
		}
	}
}

func (c *Connection) sendBytes(chID byte, msg []byte) error {
	ln := uint64(len(msg)) + 1
	err := writeUvarint(c.s, ln)
	if err == nil {
		c.s.Write([]byte{chID})
		_, err = c.s.Write(msg)
	}
	return err
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

func writeUvarint(w io.Writer, i uint64) error {
	varintbuf := make([]byte, 16)
	n := binary.PutUvarint(varintbuf, i)
	_, err := w.Write(varintbuf[:n])
	if err != nil {
		return err
	}
	return nil
}
