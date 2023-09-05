package transfer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnType(t *testing.T) {
	assert.Equal(t, "unix-seqpacket", ConnTypeUnixSeqPacket.String())
	assert.Equal(t, "unix-stream", ConnTypeUnixStream.String())
	assert.Equal(t, "invalid-socket-type", ConnType(123).String())
}
