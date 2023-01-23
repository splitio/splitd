package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/util"
	"gopkg.in/yaml.v3"
)

func main() {

	cfg, err := readConfig()
	if err != nil {
		fmt.Println("error reading config: ", err.Error())
		os.Exit(1)
	}

	logger := logging.NewLogger(&logging.LoggerOptions{
		StandardLoggerFlags: log.Ltime | log.Lshortfile,
		LogLevel: logging.LevelInfo,
	})

	splitSDK, err := sdk.New(logger, cfg.Apikey)

	errc, lShutdown, err := link.Listen(logger, splitSDK, cfg.Link.toLinkOpts()...)
	if err != nil {
		logger.Error("startup error: ", err)
		os.Exit(1)
	}

	shutdown := util.NewShutdownHandler()
	shutdown.RegisterHook(func() {
		err := lShutdown()
		if err != nil {
			logger.Error(err)
		}
	})
	defer shutdown.Wait()

	// Wait for connection to end (either gracefully of because of an error)
	err = <-errc
	if err != nil {
		logger.Error(err)
	}
}

// -- Config

const defaultConfigFN = "/etc/splitd.yaml"

type config struct {
	Apikey string `yaml:"apikey"`
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
	Type          *string `yaml:"type"`
	Address       *string `yaml:"address"`
	Serialization *string `yaml:"serialization"`
}

func (l Link) toLinkOpts() []link.Option {
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
	return opts
}

func readConfig() (*config, error) {
	cfgFN := defaultConfigFN
	if fromEnv := os.Getenv("SPLITD_CONF_FILE"); fromEnv != "" {
		cfgFN = fromEnv
	}

	var c config
	return &c, c.parse(cfgFN)
}


