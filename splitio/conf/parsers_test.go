package conf

import (
	"testing"

	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/stretchr/testify/assert"
)

func TestParseProtocol(t *testing.T) {
	pv, err := parseProtocolVersion("v1")
	assert.Nil(t, err)
	assert.Equal(t, protocol.V1, pv)

	pv, err = parseProtocolVersion("v2")
	assert.NotNil(t, err)
	assert.NotEqual(t, pv, protocol.V1)
}

func TestParseConnType(t *testing.T) {
	ct, err := parseConnType("unix-stream")
	assert.Nil(t, err)
	assert.Equal(t, transfer.ConnTypeUnixStream, ct)

	ct, err = parseConnType("unix-seqpacket")
	assert.Nil(t, err)
	assert.Equal(t, transfer.ConnTypeUnixSeqPacket, ct)

	ct, err = parseConnType("something-else")
	assert.NotNil(t, err)
	assert.NotEqual(t, transfer.ConnTypeUnixSeqPacket, ct)
	assert.NotEqual(t, transfer.ConnTypeUnixStream, ct)
}

func TestParseSerializer(t *testing.T) {
	sm, err := parseSerializer("msgpack")
	assert.Nil(t, err)
	assert.Equal(t, serializer.MsgPack, sm)

	sm, err = parseSerializer("something_esle")
	assert.NotNil(t, err)
	assert.NotEqual(t, serializer.MsgPack, sm)

}
