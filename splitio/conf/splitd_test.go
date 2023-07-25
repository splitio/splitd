package conf

import (
	"testing"
	"time"

	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/stretchr/testify/assert"
)

func TestLink(t *testing.T) {

	linkCFG := &Link{
		Type:                 ref("unix-stream"),
		Address:              ref("some/file"),
		MaxSimultaneousConns: ref(1),
		ReadTimeoutMS:        ref(2),
		WriteTimeoutMS:       ref(3),
		AcceptTimeoutMS:      ref(4),
		Serialization:        ref("msgpack"),
		BufferSize:           ref(5),
		Protocol:             ref("v1"),
	}

	expected := link.DefaultListenerOptions()
	expected.Acceptor.AcceptTimeout = 4 * time.Millisecond
	expected.Acceptor.MaxSimultaneousConnections = 1
	expected.Protocol = protocol.V1
	expected.Serialization = serializer.MsgPack
	expected.Transfer.Address = "some/file"
	expected.Transfer.ConnType = transfer.ConnTypeUnixStream
	expected.Transfer.ReadTimeout = 2 * time.Millisecond
	expected.Transfer.WriteTimeout = 3 * time.Millisecond
	expected.Transfer.BufferSize = 5
	lopts, err := linkCFG.ToListenerOpts()
	assert.Nil(t, err)
	assert.Equal(t, &expected, lopts)

	// invalid protocol
	linkCFG.Protocol = ref("sarasa")
	lopts, err = linkCFG.ToListenerOpts()
	assert.NotNil(t, err)
	assert.Nil(t, lopts)

	// invalid serialization
	linkCFG.Protocol = ref("v1") // restore valid protocol
	linkCFG.Serialization = ref("sarasa")
	lopts, err = linkCFG.ToListenerOpts()
	assert.NotNil(t, err)
	assert.Nil(t, lopts)

	// invalid conn type
	linkCFG.Serialization = ref("msgpack") // restore valid serialization mechanism
	linkCFG.Type = ref("sarasa")
	lopts, err = linkCFG.ToListenerOpts()
	assert.NotNil(t, err)
	assert.Nil(t, lopts)
}

func TestSDK(t *testing.T) {

	sdkCFG := &SDK{
		Apikey:           "some",
		LabelsEnabled:    ref(false),
		StreamingEnabled: ref(false),
		URLs: URLs{
			Auth:      ref("authURL"),
			SDK:       ref("sdkURL"),
			Events:    ref("eventsURL"),
			Streaming: ref("streamingURL"),
			Telemetry: ref("telemetryURL"),
		},
	}

	expected := conf.DefaultConfig()
	expected.StreamingEnabled = false
	expected.LabelsEnabled = false
	expected.URLs.Auth = "authURL"
	expected.URLs.SDK = "sdkURL"
	expected.URLs.Events = "eventsURL"
	expected.URLs.Streaming = "streamingURL"
	expected.URLs.Telemetry = "telemetryURL"

	assert.Equal(t, expected, sdkCFG.ToSDKConf())
}

func ref[T any](v T) *T {
	return &v
}
