package p2p

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/go-common"
	wire "github.com/tendermint/go-wire"
)

func TestPEXReactorBasic(t *testing.T) {
	dir, err := ioutil.TempDir("", "pex_reactor")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	book := NewAddrBook(dir+"addrbook.json", true)

	r := NewPEXReactor(book)

	assert.NotNil(t, r)
	assert.NotEmpty(t, r.GetChannels())
}

func TestPEXReactorAddRemovePeer(t *testing.T) {
	dir, err := ioutil.TempDir("", "pex_reactor")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	book := NewAddrBook(dir+"addrbook.json", true)

	r := NewPEXReactor(book)

	size := book.Size()
	peer := createRandomPeer(false)

	r.AddPeer(peer)
	assert.Equal(t, size+1, book.Size())

	r.RemovePeer(peer, "peer not available")
	assert.Equal(t, size, book.Size())

	outboundPeer := createRandomPeer(true)

	r.AddPeer(outboundPeer)
	assert.Equal(t, size, book.Size(), "size must not change")

	r.RemovePeer(outboundPeer, "peer not available")
	assert.Equal(t, size, book.Size(), "size must not change")
}

func TestPEXReactorRunning(t *testing.T) {
	N := 3
	switches := make([]*Switch, N)

	dir, err := ioutil.TempDir("", "pex_reactor")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	book := NewAddrBook(dir+"addrbook.json", false)

	// create switches
	for i := 0; i < N; i++ {
		switches[i] = makeSwitch(i, "127.0.0.1", "123.123.123", func(i int, sw *Switch) *Switch {
			r := NewPEXReactor(book)
			r.SetEnsurePeersPeriod(250 * time.Millisecond)
			sw.AddReactor("pex", r)
			return sw
		})
	}

	// fill the address book and add listeners
	for _, s := range switches {
		addr := NewNetAddressString(s.NodeInfo().ListenAddr)
		book.AddAddress(addr, addr)
		s.AddListener(NewDefaultListener("tcp", s.NodeInfo().ListenAddr, true))
	}

	// start switches
	for _, s := range switches {
		_, err := s.Start() // start switch and reactors
		require.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	// check peers are connected after some time
	for _, s := range switches {
		outbound, inbound, _ := s.NumPeers()
		if outbound+inbound == 0 {
			t.Errorf("%v expected to be connected to at least one peer", s.NodeInfo().ListenAddr)
		}
	}

	// stop them
	for _, s := range switches {
		s.Stop()
	}
}

func TestPEXReactorReceive(t *testing.T) {
	dir, err := ioutil.TempDir("", "pex_reactor")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	book := NewAddrBook(dir+"addrbook.json", true)

	r := NewPEXReactor(book)

	peer := createRandomPeer(false)

	size := book.Size()
	addrs := []*NetAddress{NewNetAddressString(peer.ListenAddr)}
	msg := wire.BinaryBytes(struct{ PexMessage }{&pexAddrsMessage{Addrs: addrs}})
	r.Receive(PexChannel, peer, msg)
	assert.Equal(t, size+1, book.Size())

	msg = wire.BinaryBytes(struct{ PexMessage }{&pexRequestMessage{}})
	r.Receive(PexChannel, peer, msg)
}

func TestPEXReactorAbuseFromPeer(t *testing.T) {
	dir, err := ioutil.TempDir("", "pex_reactor")
	require.Nil(t, err)
	defer os.RemoveAll(dir)
	book := NewAddrBook(dir+"addrbook.json", true)

	r := NewPEXReactor(book)
	r.SetMaxMsgCountByPeer(5)

	peer := createRandomPeer(false)

	msg := wire.BinaryBytes(struct{ PexMessage }{&pexRequestMessage{}})
	for i := 0; i < 10; i++ {
		r.Receive(PexChannel, peer, msg)
	}

	assert.True(t, r.ReachedMaxMsgCountForPeer(peer.ListenAddr))
}

func createRandomPeer(outbound bool) *Peer {
	addr := cmn.Fmt("%v.%v.%v.%v:46656", rand.Int()%256, rand.Int()%256, rand.Int()%256, rand.Int()%256)
	return &Peer{
		Key: cmn.RandStr(12),
		NodeInfo: &NodeInfo{
			ListenAddr: addr,
		},
		outbound: outbound,
		mconn:    &MConnection{RemoteAddress: NewNetAddressString(addr)},
	}
}
