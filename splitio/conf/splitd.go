package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
	sdlogging "github.com/splitio/splitd/splitio/logging"
	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	cc "github.com/splitio/splitd/splitio/util/conf"
	"gopkg.in/yaml.v3"
)

const defaultConfigFN = "/etc/splitd.yaml"

type Config struct {
	Logger Logger `yaml:"logging"`
	SDK    SDK    `yaml:"sdk"`
	Link   Link   `yaml:"link"`
}

func (c Config) String() string {
	if len(c.SDK.Apikey) > 4 {
		c.SDK.Apikey = c.SDK.Apikey[:4] + "xxxxxxx"
	}

	output, _ := json.Marshal(c)
	return string(output)
}

func (c *Config) parse(fn string) error {

	raw, err := ioutil.ReadFile(fn)
	if err != nil {
		return fmt.Errorf("error reading yaml file: %w", err)
	}

	err = yaml.Unmarshal(raw, c)
	if err != nil {
		return fmt.Errorf("error parsing yaml file: %w", err)
	}

	return nil
}

type Link struct {
	Type                 *string `yaml:"type"`
	Address              *string `yaml:"address"`
	MaxSimultaneousConns *int    `yaml:"maxSimultaneousConns"`
	ReadTimeoutMS        *int    `yaml:"readTimeoutMS"`
	WriteTimeoutMS       *int    `yaml:"writeTimeoutMS"`
	AcceptTimeoutMS      *int    `yaml:"acceptTimeoutMS"`
	Serialization        *string `yaml:"serialization"`
	BufferSize           *int    `yaml:"bufferSize"`
	Protocol             *string `yaml:"protocol"`
}

func (l *Link) ToListenerOpts() (*link.ListenerOptions, error) {
	opts := link.DefaultListenerOptions()

	var err error
	if l.Protocol != nil {
		if opts.Protocol, err = cc.ParseProtocolVersion(*l.Protocol); err != nil {
			return nil, fmt.Errorf("invalid protocol version %s", *l.Protocol)
		}
	}

	if l.Type != nil {
		if opts.Transfer.ConnType, err = cc.ParseConnType(*l.Type); err != nil {
			return nil, fmt.Errorf("invalid connection type %s", *l.Type)
		}
	}

	if l.Serialization != nil {
		if opts.Serialization, err = cc.ParseSerializer(*l.Serialization); err != nil {
			return nil, fmt.Errorf("invalid serialization %s", *l.Serialization)
		}
	}

	durationFromMS := func(i int) time.Duration { return time.Duration(i) * time.Millisecond }
	cc.SetIfNotNil(&opts.Transfer.Address, l.Address)
	cc.SetIfNotNil(&opts.Transfer.BufferSize, l.BufferSize)
	cc.SetIfNotNil(&opts.Acceptor.MaxSimultaneousConnections, l.MaxSimultaneousConns)
	cc.MapIfNotNil(&opts.Transfer.ReadTimeout, l.ReadTimeoutMS, durationFromMS)
	cc.MapIfNotNil(&opts.Transfer.WriteTimeout, l.WriteTimeoutMS, durationFromMS)
	cc.MapIfNotNil(&opts.Acceptor.AcceptTimeout, l.AcceptTimeoutMS, durationFromMS)

	return &opts, nil
}

type SDK struct {
	Apikey           string       `yaml:"apikey"`
	LabelsEnabled    *bool        `yaml:"labelsEnabled"`
	StreamingEnabled *bool        `yaml:"streamingEnabled"`
	URLs             URLs         `yaml:"urls"`
	FeatureFlags     FeatureFlags `yaml:"featureFlags"`
	Impressions      Impressions  `yaml:"impressions"`
}

type FeatureFlags struct {
	SplitNotificationQueueSize   *int `yaml:"splitNotificationQueueSize"`
	SplitRefreshRateSeconds      *int `yaml:"splitRefreshSeconds"`
	SegmentNotificationQueueSize *int `yaml:"segmentNotificationQueueSize"`
	SegmentRefreshRateSeconds    *int `yaml:"segmentRefreshSeconds"`
	SegmentUpdateWorkers         *int `yaml:"segmentUpdateWorkers"`
	SegmentUpdateQueueSize       *int `yaml:"segmentUpdateQueueSize"`
}

type Impressions struct {
	Mode                    *string `yaml:"mode"`
	RefreshRateSeconds      *int    `yaml:"refreshRateSeconds"`
	CountRefreshRateSeconds *int    `yaml:"countRefreshRateSeconds"`
	QueueSize               *int    `yaml:"queueSize"`
	ObserverSize            *int    `yaml:"observerSize"`
	Watermark               *int    `yaml:"watermark"`
}

func (s *SDK) ToSDKConf() *sdkConf.Config {
	cfg := sdkConf.DefaultConfig()
	durationFromSeconds := func(seconds int) time.Duration { return time.Duration(seconds) * time.Second }
	cc.SetIfNotNil(&cfg.LabelsEnabled, s.LabelsEnabled)
	cc.SetIfNotNil(&cfg.StreamingEnabled, s.StreamingEnabled)
	cc.SetIfNotEmpty(&cfg.Splits.UpdateBufferSize, s.FeatureFlags.SplitNotificationQueueSize)
	cc.MapIfNotNil(&cfg.Splits.SyncPeriod, s.FeatureFlags.SplitRefreshRateSeconds, durationFromSeconds)
	cc.SetIfNotEmpty(&cfg.Segments.UpdateBufferSize, s.FeatureFlags.SegmentNotificationQueueSize)
	cc.SetIfNotEmpty(&cfg.Segments.QueueSize, s.FeatureFlags.SegmentUpdateQueueSize)
	cc.SetIfNotEmpty(&cfg.Segments.WorkerCount, s.FeatureFlags.SegmentUpdateWorkers)
	cc.MapIfNotNil(&cfg.Segments.SyncPeriod, s.FeatureFlags.SegmentRefreshRateSeconds, durationFromSeconds)
	cc.SetIfNotEmpty(&cfg.Impressions.Mode, s.Impressions.Mode)
	cc.SetIfNotEmpty(&cfg.Impressions.ObserverSize, s.Impressions.ObserverSize)
	cc.SetIfNotEmpty(&cfg.Impressions.QueueSize, s.Impressions.QueueSize)
	cc.MapIfNotNil(&cfg.Impressions.SyncPeriod, s.Impressions.RefreshRateSeconds, durationFromSeconds)
	cc.MapIfNotNil(&cfg.Impressions.CountSyncPeriod, s.Impressions.CountRefreshRateSeconds, durationFromSeconds)
	s.URLs.updateSDKConfURLs(&cfg.URLs)
	return cfg
}

type URLs struct {
	Auth      *string `yaml:"auth"`
	SDK       *string `yaml:"sdk"`
	Events    *string `yaml:"events"`
	Streaming *string `yaml:"streaming"`
	Telemetry *string `yaml:"telemetry"`
}

func (u *URLs) updateSDKConfURLs(dst *sdkConf.URLs) {
	cc.SetIfNotNil(&dst.SDK, u.SDK)
	cc.SetIfNotNil(&dst.Events, u.Events)
	cc.SetIfNotNil(&dst.Auth, u.Auth)
	cc.SetIfNotNil(&dst.Streaming, u.Streaming)
	cc.SetIfNotNil(&dst.Telemetry, u.Telemetry)
}

type Logger struct {
	Level                   *string `yaml:"level"`
	Output                  *string `yaml:"file"`
	RotationMaxFiles        *int    `yaml:"rotationMaxFiles"`
	RotationMaxBytesPerFile *int    `yaml:"rotationMaxBytesPerFile"`
}

func (l *Logger) ToLoggerOptions() (*logging.LoggerOptions, error) {

	writer, err := sdlogging.GetWriter(l.Output, l.RotationMaxFiles, l.RotationMaxBytesPerFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing logger options: %w", err)
	}

	opts := &logging.LoggerOptions{
		LogLevel:            logging.LevelError,
		StandardLoggerFlags: log.Ltime | log.Lshortfile,
		ErrorWriter:         writer,
		WarningWriter:       writer,
		InfoWriter:          writer,
		DebugWriter:         writer,
		VerboseWriter:       writer,
	}

	if l.Level != nil {
		opts.LogLevel = logging.Level(strings.ToUpper(*l.Level))
	}

	return opts, nil
}

func ReadConfig() (*Config, error) {
	cfgFN := defaultConfigFN
	if fromEnv := os.Getenv("SPLITD_CONF_FILE"); fromEnv != "" {
		cfgFN = fromEnv
	}

	var c Config
	return &c, c.parse(cfgFN)
}
