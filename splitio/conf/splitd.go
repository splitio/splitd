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
	"github.com/splitio/splitd/splitio/common/lang"
	"github.com/splitio/splitd/splitio/link"
	sdlogging "github.com/splitio/splitd/splitio/logging"
	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigFN   = "/etc/splitd.yaml"
	apikeyPlaceHolder = "<server-side-apitoken>"
	defaultLogLevel   = "error"
	defaultLogOutput  = "/dev/stdout"
)

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

func (c *Config) PopulateWithDefaults() {
	c.SDK.PopulateWithDefaults()
	c.Link.PopulateWithDefaults()
	c.Logger.PopulateWithDefaults()
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

func (l *Link) PopulateWithDefaults() {
	linkOpts := link.DefaultListenerOptions()
	l.Address = lang.Ref(linkOpts.Transfer.Address)
	l.Type = lang.Ref(linkOpts.Transfer.ConnType.String())
	l.ReadTimeoutMS = lang.Ref(int(linkOpts.Transfer.ReadTimeout.Milliseconds()))
	l.WriteTimeoutMS = lang.Ref(int(linkOpts.Transfer.WriteTimeout.Milliseconds()))
	l.AcceptTimeoutMS = lang.Ref(int(linkOpts.Acceptor.AcceptTimeout.Milliseconds()))
	l.BufferSize = lang.Ref(linkOpts.Transfer.BufferSize)
	l.MaxSimultaneousConns = lang.Ref(linkOpts.Acceptor.MaxSimultaneousConnections)
	l.Protocol = lang.Ref(linkOpts.Protocol.String())
	l.Serialization = lang.Ref(linkOpts.Serialization.String())
}

func (l *Link) ToListenerOpts() (*link.ListenerOptions, error) {
	opts := link.DefaultListenerOptions()

	var err error
	if l.Protocol != nil {
		if opts.Protocol, err = parseProtocolVersion(*l.Protocol); err != nil {
			return nil, fmt.Errorf("invalid protocol version %s", *l.Protocol)
		}
	}

	if l.Type != nil {
		if opts.Transfer.ConnType, err = parseConnType(*l.Type); err != nil {
			return nil, fmt.Errorf("invalid connection type %s", *l.Type)
		}
	}

	if l.Serialization != nil {
		if opts.Serialization, err = parseSerializer(*l.Serialization); err != nil {
			return nil, fmt.Errorf("invalid serialization %s", *l.Serialization)
		}
	}

	durationFromMS := func(i int) time.Duration { return time.Duration(i) * time.Millisecond }
	lang.SetIfNotNil(&opts.Transfer.Address, l.Address)
	lang.SetIfNotNil(&opts.Transfer.BufferSize, l.BufferSize)
	lang.SetIfNotNil(&opts.Acceptor.MaxSimultaneousConnections, l.MaxSimultaneousConns)
	lang.MapIfNotNil(&opts.Transfer.ReadTimeout, l.ReadTimeoutMS, durationFromMS)
	lang.MapIfNotNil(&opts.Transfer.WriteTimeout, l.WriteTimeoutMS, durationFromMS)
	lang.MapIfNotNil(&opts.Acceptor.AcceptTimeout, l.AcceptTimeoutMS, durationFromMS)

	return &opts, nil
}

type SDK struct {
	Apikey           string       `yaml:"apikey"`
	LabelsEnabled    *bool        `yaml:"labelsEnabled"`
	StreamingEnabled *bool        `yaml:"streamingEnabled"`
	URLs             URLs         `yaml:"urls"`
	FeatureFlags     FeatureFlags `yaml:"featureFlags"`
	Impressions      Impressions  `yaml:"impressions"`
	Events           Events       `yank:"events"`
}

func (s *SDK) PopulateWithDefaults() {
	cfg := sdkConf.DefaultConfig()
	s.Apikey = apikeyPlaceHolder
	s.LabelsEnabled = lang.Ref(cfg.LabelsEnabled)
	s.StreamingEnabled = lang.Ref(cfg.StreamingEnabled)
	s.URLs.PopulateWithDefaults()
	s.FeatureFlags.PopulateWithDefaults()
	s.Impressions.PopulateWithDefaults()
	s.Events.PopulateWithDefaults()
}

type FeatureFlags struct {
	SplitNotificationQueueSize   *int `yaml:"splitNotificationQueueSize"`
	SplitRefreshRateSeconds      *int `yaml:"splitRefreshSeconds"`
	SegmentNotificationQueueSize *int `yaml:"segmentNotificationQueueSize"`
	SegmentRefreshRateSeconds    *int `yaml:"segmentRefreshSeconds"`
	SegmentWorkerCount           *int `yaml:"segmentUpdateWorkers"`
	SegmentWorkerBufferSize      *int `yaml:"segmentUpdateQueueSize"`
}

func (f *FeatureFlags) PopulateWithDefaults() {
	ffOpts := sdkConf.DefaultConfig()
	f.SegmentNotificationQueueSize = lang.Ref(ffOpts.Segments.UpdateBufferSize)
	f.SegmentRefreshRateSeconds = lang.Ref(int(ffOpts.Segments.SyncPeriod.Seconds()))
	f.SegmentWorkerBufferSize = lang.Ref(ffOpts.Segments.QueueSize)
	f.SegmentWorkerCount = lang.Ref(ffOpts.Segments.WorkerCount)
	f.SplitNotificationQueueSize = lang.Ref(ffOpts.Splits.UpdateBufferSize)
	f.SplitRefreshRateSeconds = lang.Ref(int(ffOpts.Splits.SyncPeriod.Seconds()))
}

type Impressions struct {
	Mode                    *string `yaml:"mode"`
	RefreshRateSeconds      *int    `yaml:"refreshRateSeconds"`
	CountRefreshRateSeconds *int    `yaml:"countRefreshRateSeconds"`
	QueueSize               *int    `yaml:"queueSize"`
	ObserverSize            *int    `yaml:"observerSize"`
	Watermark               *int    `yaml:"watermark,omitempty"` // TODO(mredolatti) remove omitempty when fully implemented
}

func (i *Impressions) PopulateWithDefaults() {
	cfg := sdkConf.DefaultConfig().Impressions
	i.CountRefreshRateSeconds = lang.Ref(int(cfg.CountSyncPeriod.Seconds()))
	i.Mode = lang.Ref(cfg.Mode)
	i.ObserverSize = lang.Ref(cfg.ObserverSize)
	i.RefreshRateSeconds = lang.Ref(int(cfg.SyncPeriod.Seconds()))
	i.QueueSize = lang.Ref(cfg.QueueSize)
}

type Events struct {
	RefreshRateSeconds *int `yaml:"refreshRateSeconds"`
	QueueSize          *int `yaml:"queueSize"`
	Watermark          *int `yaml:"watermark,omitempty"` // TODO(mredolatti) remove omitempty when fully implemented
}

func (e *Events) PopulateWithDefaults() {
	cfg := sdkConf.DefaultConfig().Events
	e.RefreshRateSeconds = lang.Ref(int(cfg.SyncPeriod.Seconds()))
	e.QueueSize = lang.Ref(cfg.QueueSize)
}

func (s *SDK) ToSDKConf() *sdkConf.Config {
	cfg := sdkConf.DefaultConfig()
	durationFromSeconds := func(seconds int) time.Duration { return time.Duration(seconds) * time.Second }
	lang.SetIfNotNil(&cfg.LabelsEnabled, s.LabelsEnabled)
	lang.SetIfNotNil(&cfg.StreamingEnabled, s.StreamingEnabled)
	lang.SetIfNotEmpty(&cfg.Splits.UpdateBufferSize, s.FeatureFlags.SplitNotificationQueueSize)
	lang.MapIfNotNil(&cfg.Splits.SyncPeriod, s.FeatureFlags.SplitRefreshRateSeconds, durationFromSeconds)
	lang.SetIfNotEmpty(&cfg.Segments.UpdateBufferSize, s.FeatureFlags.SegmentNotificationQueueSize)
	lang.SetIfNotEmpty(&cfg.Segments.QueueSize, s.FeatureFlags.SegmentWorkerBufferSize)
	lang.SetIfNotEmpty(&cfg.Segments.WorkerCount, s.FeatureFlags.SegmentWorkerCount)
	lang.MapIfNotNil(&cfg.Segments.SyncPeriod, s.FeatureFlags.SegmentRefreshRateSeconds, durationFromSeconds)
	lang.SetIfNotEmpty(&cfg.Impressions.Mode, s.Impressions.Mode)
	lang.SetIfNotEmpty(&cfg.Impressions.ObserverSize, s.Impressions.ObserverSize)
	lang.SetIfNotEmpty(&cfg.Impressions.QueueSize, s.Impressions.QueueSize)
	lang.MapIfNotNil(&cfg.Impressions.SyncPeriod, s.Impressions.RefreshRateSeconds, durationFromSeconds)
	lang.MapIfNotNil(&cfg.Impressions.CountSyncPeriod, s.Impressions.CountRefreshRateSeconds, durationFromSeconds)
	lang.SetIfNotEmpty(&cfg.Events.QueueSize, s.Events.QueueSize)
	lang.MapIfNotNil(&cfg.Events.SyncPeriod, s.Events.RefreshRateSeconds, durationFromSeconds)
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
	lang.SetIfNotNil(&dst.SDK, u.SDK)
	lang.SetIfNotNil(&dst.Events, u.Events)
	lang.SetIfNotNil(&dst.Auth, u.Auth)
	lang.SetIfNotNil(&dst.Streaming, u.Streaming)
	lang.SetIfNotNil(&dst.Telemetry, u.Telemetry)
}

func (u *URLs) PopulateWithDefaults() {
	cfg := sdkConf.DefaultConfig().URLs
	u.Auth = lang.Ref(cfg.Auth)
	u.Events = lang.Ref(cfg.Events)
	u.SDK = lang.Ref(cfg.SDK)
	u.Streaming = lang.Ref(cfg.Streaming)
	u.Telemetry = lang.Ref(cfg.Telemetry)
}

type Logger struct {
	Level                   *string `yaml:"level"`
	Output                  *string `yaml:"output"`
	RotationMaxFiles        *int    `yaml:"rotationMaxFiles"`
	RotationMaxBytesPerFile *int    `yaml:"rotationMaxBytesPerFile"`
}

func (l *Logger) PopulateWithDefaults() {
	l.Level = lang.Ref(defaultLogLevel)
	l.Output = lang.Ref(defaultLogOutput)
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
