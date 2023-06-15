package main

import (
	"fmt"
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

	logger := logging.NewLogger(cfg.Logger.ToLoggerOptions())

	splitSDK, err := sdk.New(logger, cfg.SDK.Apikey, cfg.SDK.ToSDKConf()...)
    exitOnErr("sdk initialization", err)

	errc, lShutdown, err := link.Listen(logger, splitSDK, cfg.Link.ToLinkOpts()...)
    exitOnErr("rpc listener setup", err)

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
    exitOnErr("shutdown: ", err)
}

func exitOnErr(ctxStr string, err error) {
	if err != nil {
        fmt.Printf("%s: startup error: %s\n", ctxStr, err.Error())
		os.Exit(1)
	}

}
