package p2p

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendRecvBytes(t *testing.T) {
	b := &bytes.Buffer{}
	old := uint64(20)
	err := writeUvarint(b, old)
	assert.NoError(t, err)
	r := bytes.NewReader(b.Bytes())
	x, err := readUvarint(r)
	assert.Equal(t, x, old)
	assert.NoError(t, err)
}
