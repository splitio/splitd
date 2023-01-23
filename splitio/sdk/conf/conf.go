package conf

import (
	"time"

	"github.com/splitio/go-split-commons/v4/conf"
)

type Config struct {
	LabelsEnabled    bool
	StreamingEnabled bool
	Splits           Splits
	Segments         Segments
	Impressions      Impressions
	URLs             URLs
}

func (c *Config) ParseOptions(options []Option) error {
	for _, apply := range options {
		err := apply(c)
		if err != nil {
			return err
		}
	}
	return nil
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
	d.SegmentsRefreshRate = int(c.Segments.SyncPeriod.Seconds())
	d.StreamingEnabled = c.StreamingEnabled
	d.AuthServiceURL = c.URLs.Auth
	d.SdkURL = c.URLs.SDK
	d.EventsURL = c.URLs.Events
	d.StreamingServiceURL = c.URLs.Streaming
	d.TelemetryServiceURL = c.URLs.Telemetry
	// TODO(update with custom opts)
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
			SyncPeriod:      5 * time.Second,
			CountSyncPeriod: 5 * time.Second,
			PostConcurrency: 1,
		},
		URLs: URLs{
			Auth:      "https://auth.split.io",
			SDK:       "https://sdk.split.io/api",
			Events:    "https://events.split.io/api",
			Streaming: "https://streaming.split.io/sse",
			Telemetry: "https://telemetry.split.io/api/v1",
		},
	}
}

type Option func(c *Config) error

func WithLabelsEnabled(v bool) Option {
	return func(c *Config) error { c.LabelsEnabled = v; return nil }
}

func WithStreamingEnabled(v bool) Option {
	return func(c *Config) error { c.StreamingEnabled = v; return nil }
}

func WithAuthURL(v string) Option {
	return func(c *Config) error { c.URLs.Auth = v; return nil }
}

func WithSDKURL(v string) Option {
	return func(c *Config) error { c.URLs.SDK = v; return nil }
}

func WithEventsURL(v string) Option {
	return func(c *Config) error { c.URLs.Events = v; return nil }
}

func WithStreamingURL(v string) Option {
	return func(c *Config) error { c.URLs.Streaming = v; return nil }
}


func WithTelemetryURL(v string) Option {
	return func(c *Config) error { c.URLs.Telemetry = v; return nil }
}
