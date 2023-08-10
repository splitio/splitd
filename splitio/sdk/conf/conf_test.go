package conf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSDKConf(t *testing.T) {
	dc := DefaultConfig()
	dc.Impressions.Mode = "debug"
	dc.Impressions.SyncPeriod = 1 * time.Minute
	warns := dc.Normalize()
	assert.Equal(t, warns, []string{
		"only `optimized` impressions mode supported currently. ignoring user config",
		"minimum impressions refresh rate is 30 min. ignoring user config",
	})
	assert.Equal(t, "optimized", dc.Impressions.Mode)
	assert.Equal(t, 30*time.Minute, dc.Impressions.SyncPeriod)


	adv := dc.ToAdvancedConfig()
	assert.Equal(t, 30, adv.HTTPTimeout)
	assert.Equal(t, dc.Segments.QueueSize, adv.SegmentQueueSize)
	assert.Equal(t, dc.Segments.WorkerCount, adv.SegmentWorkers)
	assert.Equal(t, dc.URLs.SDK, adv.SdkURL)
	assert.Equal(t, dc.URLs.Events, adv.EventsURL)
	assert.Equal(t, dc.URLs.Telemetry, adv.TelemetryServiceURL)
	// assert.Equal(t, TODO, adv.EventsBulkSize)
	// assert.Equal(t, TODO, adv.EventsQueueSize)
	assert.Equal(t, dc.Impressions.QueueSize, adv.ImpressionsQueueSize)
	// assert.Equal(t, TODO, adv.ImpressionsBulkSize)
	assert.Equal(t, dc.StreamingEnabled, adv.StreamingEnabled)
	assert.Equal(t, dc.URLs.Auth, adv.AuthServiceURL)
	assert.Equal(t, dc.URLs.Streaming, adv.StreamingServiceURL)
	assert.Equal(t, int64(dc.Splits.UpdateBufferSize), adv.SplitUpdateQueueSize)
	assert.Equal(t, int64(dc.Segments.UpdateBufferSize), adv.SegmentUpdateQueueSize)
	assert.Equal(t, int(dc.Splits.SyncPeriod.Seconds()), adv.SplitsRefreshRate)
	assert.Equal(t, int(dc.Segments.SyncPeriod.Seconds()), adv.SegmentsRefreshRate)
}
