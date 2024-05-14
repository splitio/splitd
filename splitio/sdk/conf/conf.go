package conf

import (
	"time"

	"github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/flagsets"
	"github.com/splitio/go-split-commons/v6/service/api/specs"
)

const (
	defaultImpressionsMode        = "optimized"
	minimumImpressionsRefreshRate = 30 * time.Minute
)

type Config struct {
	LabelsEnabled    bool
	StreamingEnabled bool
	Splits           Splits
	Segments         Segments
	Impressions      Impressions
	Events           Events
	URLs             URLs
	FlagSetsFilter   []string
}

type Splits struct {
	SyncPeriod       time.Duration
	UpdateBufferSize int
}

type Segments struct {
	SyncPeriod       time.Duration
	UpdateBufferSize int
	WorkerCount      int
	QueueSize        int
}

type Impressions struct {
	Mode            string
	ObserverSize    int
	QueueSize       int
	SyncPeriod      time.Duration
	CountSyncPeriod time.Duration
	PostConcurrency int
}

type Events struct {
	QueueSize       int
	SyncPeriod      time.Duration
	PostConcurrency int
}

type URLs struct {
	Auth      string
	SDK       string
	Events    string
	Streaming string
	Telemetry string
}

func (c *Config) ToAdvancedConfig() *conf.AdvancedConfig {
	d := conf.GetDefaultAdvancedConfig()

	d.SplitsRefreshRate = int(c.Splits.SyncPeriod.Seconds())
	d.SplitUpdateQueueSize = int64(c.Splits.UpdateBufferSize)
	d.SegmentsRefreshRate = int(c.Segments.SyncPeriod.Seconds())
	d.SegmentQueueSize = c.Segments.QueueSize
	d.SegmentUpdateQueueSize = int64(c.Segments.UpdateBufferSize)
	d.SegmentWorkers = c.Segments.WorkerCount
	d.StreamingEnabled = c.StreamingEnabled
	d.FlagSetsFilter = c.FlagSetsFilter

	d.AuthServiceURL = c.URLs.Auth
	d.SdkURL = c.URLs.SDK
	d.EventsURL = c.URLs.Events
	d.StreamingServiceURL = c.URLs.Streaming
	d.TelemetryServiceURL = c.URLs.Telemetry

	d.ImpressionsQueueSize = c.Impressions.QueueSize
	d.AuthSpecVersion = specs.FLAG_V1_1
	d.FlagsSpecVersion = specs.FLAG_V1_1

	return &d
}

func DefaultConfig() *Config {
	return &Config{
		LabelsEnabled:    true,
		StreamingEnabled: true,
		Splits: Splits{
			SyncPeriod:       30 * time.Second,
			UpdateBufferSize: 5000,
		},
		Segments: Segments{
			SyncPeriod:       60 * time.Second,
			WorkerCount:      20,
			QueueSize:        500,
			UpdateBufferSize: 5000,
		},
		Impressions: Impressions{
			Mode:            "optimized",
			ObserverSize:    500000,
			QueueSize:       8192,
			SyncPeriod:      30 * time.Minute,
			CountSyncPeriod: 60 * time.Minute,
			PostConcurrency: 1,
		},
		Events: Events{
			QueueSize:       8192,
			SyncPeriod:      1 * time.Minute,
			PostConcurrency: 1,
		},
		URLs: URLs{
			Auth:      "https://auth.split.io",
			SDK:       "https://sdk.split.io/api",
			Events:    "https://events.split.io/api",
			Streaming: "https://streaming.split.io/sse",
			Telemetry: "https://telemetry.split.io/api/v1",
		},
		FlagSetsFilter: []string{},
	}
}

func (c *Config) Normalize() []string {
	var warnings []string
	if c.Impressions.Mode == "optimized" && c.Impressions.SyncPeriod < minimumImpressionsRefreshRate {
		warnings = append(warnings, "minimum impressions refresh rate is 30 min. ignoring user config")
		c.Impressions.SyncPeriod = minimumImpressionsRefreshRate
	}

	// Sanitize flagsets and append erros into warnings for logging purposes
	sanitizedFlagSets, warns := flagsets.SanitizeMany(c.FlagSetsFilter)
	if len(warns) != 0 {
		for _, err := range warns {
			if errType, ok := err.(dtos.FlagSetValidatonError); ok {
				warnings = append(warnings, errType.Message)
			}
		}
	}
	c.FlagSetsFilter = sanitizedFlagSets

	return warnings
}
