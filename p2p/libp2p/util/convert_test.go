package util

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/multiformats/go-multiaddr"
)

func TestMultiaddr2DialString(t *testing.T) {
	str := "/ip4/127.0.0.1/udp/1234"
	ma, _ := multiaddr.NewMultiaddr(str)
	assert.Equal(t, Multiaddr2DialString(ma), "127.0.0.1:1234")
}
