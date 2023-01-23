package main

import (
	"fmt"
	"log"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/conf"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/util"
)

func main() {

	cfg, err := conf.ReadConfig()
	if err != nil {
		fmt.Println("error reading config: ", err.Error())
		os.Exit(1)
	}

	logger := logging.NewLogger(&logging.LoggerOptions{
		StandardLoggerFlags: log.Ltime | log.Lshortfile,
		LogLevel: logging.LevelInfo,
	})

	splitSDK, err := sdk.New(logger, cfg.SDK.Apikey, cfg.SDK.ToSDKConf()...)

	errc, lShutdown, err := link.Listen(logger, splitSDK, cfg.Link.ToLinkOpts()...)
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
