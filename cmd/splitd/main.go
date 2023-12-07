package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
	"github.com/splitio/splitd/splitio/conf"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/util"

	"github.com/splitio/splitd/splitio/provisional/profiler"
)

func main() {

	printHeader()

	cfg, err := conf.ReadConfig()
	if err != nil {
		fmt.Println("error reading config: ", err.Error())
		os.Exit(1)
	}
	handleFlags(cfg)

	loggerCfg, err := cfg.Logger.ToLoggerOptions()
	exitOnErr("logging setup", err)
	logger := logging.NewLogger(loggerCfg)

	splitSDK, err := sdk.New(logger, cfg.SDK.Apikey, cfg.SDK.ToSDKConf())
	exitOnErr("sdk initialization", err)

	linkCFG, err := cfg.Link.ToListenerOpts()
	exitOnErr("link config", err)

	errc, lShutdown, err := link.Listen(logger, splitSDK, linkCFG)
	exitOnErr("rpc listener setup", err)

	shutdown := util.NewShutdownHandler()
	shutdown.RegisterHook(func() {
		err := lShutdown()
		if err != nil {
			logger.Error("error shutting down listener: ", err.Error())
		}
		splitSDK.Shutdown() // evict pending impressions & events
	})
	defer shutdown.Wait()

	if pc := cfg.Debug.Profiling; pc.Enable {
		go func() {
			p := profiler.New(pc.Host, pc.Port)
			if err := p.ListenAndServe(); err != nil {
				panic(err.Error())
			}
		}()
	}

	// Wait for connection to end (either gracefully of because of an error)
	err = <-errc
	exitOnErr("shutdown: ", err)
}

func printHeader() {
	fmt.Println(splitio.ASCILogo)
	fmt.Printf("Splitd Agent - Version %s - build [%s] (2023)\n\n", splitio.Version, splitio.CommitSHA)
}

func handleFlags(cfg *conf.Config) {
	printConf := flag.Bool("outputConfig", false, "print config (with partially obfuscated apikey)")
	flag.Parse()
	if *printConf {
		fmt.Printf("\nConfig: %s\n", cfg)
		os.Exit(0)
	}
}

func exitOnErr(ctxStr string, err error) {
	if err != nil {
		fmt.Printf("%s: startup error: %s\n", ctxStr, err.Error())
		os.Exit(1)
	}

}
