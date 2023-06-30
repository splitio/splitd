package link

import (
	"testing"
	"time"

	"github.com/splitio/splitd/splitio/link/common"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	var linkOpts Options
	err := linkOpts.populate([]Option{
		WithAcceptTimeoutMs(5),
		WithAddress("some_address"),
		WithBufSize(123),
		WithMaxSimultaneousConns(234),
		WithProtocol("v1"),
		WithReadTimeoutMs(6),
		WithSerialization("msgpack"),
		WithSockType("unix-stream"),
		WithWriteTimeoutMs(7),
	})
	assert.Nil(t, err)

	var transferOpts transfer.Options
	transferOpts.Parse(linkOpts.forTransfer())
	assert.Equal(t, 5*time.Millisecond, transferOpts.AcceptTimeout)
	assert.Equal(t, 6*time.Millisecond, transferOpts.ReadTimeout)
	assert.Equal(t, 7*time.Millisecond, transferOpts.WriteTimeout)
	assert.Equal(t, 123, transferOpts.BufferSize)
	assert.Equal(t, transfer.ConnTypeUnixStream, transferOpts.ConnType)
	assert.Equal(t, 234, transferOpts.MaxSimultaneousConnections)

	var hlOpts common.Opts
	hlOpts.Parse(linkOpts.forApp())
	assert.Equal(t, protocol.V1, hlOpts.ProtoV)
	assert.Equal(t, serializer.MsgPack, hlOpts.Serial)
}
