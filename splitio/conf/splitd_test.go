package conf

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := Config{SDK: SDK{Apikey: "someVeryLongApikey"}}
	assert.Contains(t, cfg.String(), "somexxxxxxx")

	_, filename, _, _ := runtime.Caller(0)
	parts := strings.Split(filename, string(filepath.Separator))
	dir := strings.Join(parts[:len(parts)-3], string(filepath.Separator))

	cfg = Config{}
	assert.Nil(t, cfg.parse(dir+string(filepath.Separator)+"splitd.yaml.tpl"))
	assert.Equal(t, Config{
		Logger: Logger{Level: ref("ERROR")},
		SDK: SDK{
			Apikey: "YOUR_API_KEY",
			URLs: URLs{
				Auth:      ref("https://auth.split.io"),
				SDK:       ref("https://sdk.split.io/api"),
				Events:    ref("https://events.split.io/api"),
				Streaming: ref("https://streaming.split.io/sse"),
				Telemetry: ref("https://telemetry.split.io/api/v1"),
			},
		},
		Link: Link{
			Type:          ref("unix-seqpacket"),
			Address:       ref("/var/run/splitd.sock"),
			Serialization: ref("msgpack"),
		},
	}, cfg)

	assert.Error(t, cfg.parse("someNonexistantFile"))
	assert.Error(t, cfg.parse(dir+string(filepath.Separator)+"Makefile"))

	os.Setenv("SPLITD_CONF_FILE", dir+string(filepath.Separator)+"splitd.yaml.tpl")
	newCfg, err := ReadConfig()
	assert.Nil(t, err)
	assert.NotNil(t, newCfg)

}

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
