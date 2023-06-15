package conf

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/sdk/conf"
	"gopkg.in/yaml.v3"
)

const defaultConfigFN = "/etc/splitd.yaml"

type config struct {
	Logger Logger `yaml:"logging"`
	SDK    SDK    `yaml:"sdk"`
	Link   Link   `yaml:"link"`
}

func (c *config) parse(fn string) error {

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
	Serialization        *string `yaml:"serialization"`
	MaxSimultaneousConns *int    `yaml:"maxSimultaneousConns"`
}

func (l *Link) ToLinkOpts() []link.Option {
	var opts []link.Option

	if l.Type != nil {
		opts = append(opts, link.WithSockType(*l.Type))
	}

	if l.Address != nil {
		opts = append(opts, link.WithAddress(*l.Address))
	}

	if l.Serialization != nil {
		opts = append(opts, link.WithSerialization(*l.Serialization))
	}

    if l.MaxSimultaneousConns != nil {
        opts = append(opts, link.WithMaxSimultaneousConns(*l.MaxSimultaneousConns))
    }

	return opts
}

type SDK struct {
	Apikey           string `yaml:"apikey"`
	LabelsEnabled    *bool  `yaml:"labelsEnabled"`
	StreamingEnabled *bool  `yaml:"streamingEnabled"`
	URLs             URLs   `yaml:"urls"`
}

func (s *SDK) ToSDKConf() []conf.Option {
	var opts []conf.Option
	if s.LabelsEnabled != nil {
		opts = append(opts, conf.WithLabelsEnabled(*s.LabelsEnabled))
	}
	if s.StreamingEnabled != nil {
		opts = append(opts, conf.WithStreamingEnabled(*s.StreamingEnabled))
	}
	opts = append(opts, s.URLs.ToSDKConf()...)
	return opts

}

type URLs struct {
	Auth      *string `yaml:"auth"`
	SDK       *string `yaml:"sdk"`
	Events    *string `yaml:"events"`
	Streaming *string `yaml:"streaming"`
	Telemetry *string `yaml:"telemetry"`
}

func (u *URLs) ToSDKConf() []conf.Option {
	var opts []conf.Option
	if u.Auth != nil {
		opts = append(opts, conf.WithAuthURL(*u.Auth))
	}
	if u.SDK != nil {
		opts = append(opts, conf.WithSDKURL(*u.SDK))
	}
	if u.Events != nil {
		opts = append(opts, conf.WithEventsURL(*u.Events))
	}
	if u.Streaming != nil {
		opts = append(opts, conf.WithStreamingURL(*u.Streaming))
	}
	if u.Telemetry != nil {
		opts = append(opts, conf.WithTelemetryURL(*u.Telemetry))
	}
	return opts

}

type Logger struct {
	Level *string `yaml:"level"`
}

func (l *Logger) ToLoggerOptions() *logging.LoggerOptions {

	opts := &logging.LoggerOptions{
		LogLevel:            logging.LevelError,
		StandardLoggerFlags: log.Ltime | log.Lshortfile,
	}

	if l.Level != nil {
		opts.LogLevel = logging.Level(strings.ToUpper(*l.Level))
	}

	return opts
}

func ReadConfig() (*config, error) {
	cfgFN := defaultConfigFN
	if fromEnv := os.Getenv("SPLITD_CONF_FILE"); fromEnv != "" {
		cfgFN = fromEnv
	}

	var c config
	return &c, c.parse(cfgFN)
}
