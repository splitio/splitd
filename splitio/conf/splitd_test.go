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
		FeatureFlags: FeatureFlags{
			SplitNotificationQueueSize:   ref(1),
			SplitRefreshRateSeconds:      ref(2),
			SegmentNotificationQueueSize: ref(3),
			SegmentRefreshRateSeconds:    ref(4),
			SegmentUpdateWorkers:         ref(5),
			SegmentUpdateQueueSize:       ref(6),
		},
		Impressions: Impressions{
			Mode:                    ref("optimized"),
			RefreshRateSeconds:      ref(1),
			CountRefreshRateSeconds: ref(2),
			QueueSize:               ref(3),
			ObserverSize:            ref(4),
			Watermark:               ref(5),
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
	expected.Splits.UpdateBufferSize = 1
	expected.Splits.SyncPeriod = 2 * time.Second
    expected.Segments.UpdateBufferSize = 3
    expected.Segments.SyncPeriod = 4 * time.Second
    expected.Segments.WorkerCount = 5
    expected.Segments.QueueSize = 6
    expected.Impressions.Mode = "optimized"
    expected.Impressions.SyncPeriod = 1 * time.Second
    expected.Impressions.CountSyncPeriod = 2 * time.Second
    expected.Impressions.QueueSize = 3
    expected.Impressions.ObserverSize = 4
	assert.Equal(t, expected, sdkCFG.ToSDKConf())
}

func ref[T any](v T) *T {
	return &v
}
