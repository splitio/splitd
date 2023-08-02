package main

import (
	"fmt"
	"os"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/conf"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/client/types"
	"github.com/splitio/splitd/splitio/util"
)

func main() {

	args, err := conf.ParseCliArgs()
	if err != nil {
		fmt.Println("error parsing arguments: ", err.Error())
		os.Exit(1)
	}

	linkOpts, err := args.LinkOpts()
	if err != nil {
		fmt.Println("error building options from arguments: ", err.Error())
		os.Exit(1)
	}

	logLevel := logging.Level(args.LogLevel)
	logger := logging.NewLogger(&logging.LoggerOptions{
		LogLevel:      logLevel,
		ErrorWriter:   os.Stderr,
		WarningWriter: os.Stderr,
		InfoWriter:    os.Stderr,
		DebugWriter:   os.Stderr,
		VerboseWriter: os.Stderr,
	})

	c, err := link.Consumer(logger, linkOpts)
	if err != nil {
		logger.Error("error creating client wrapper: ", err)
		os.Exit(2)
	}

	shutdown := util.NewShutdownHandler()
	shutdown.RegisterHook(func() {
		err := c.Shutdown()
		if err != nil {
			logger.Error(err.Error())
		}
	})
	defer shutdown.TriggerAndWait()

	before := time.Now()
	result, err := executeCall(c, args)
	logger.Debug(fmt.Sprintf("took: %d\n", time.Since(before).Microseconds()))
	if err != nil {
		logger.Error("error executing call: ", err.Error())
		os.Exit(3)
	}

	fmt.Println(result)
}

func executeCall(c types.ClientInterface, a *conf.CliArgs) (string, error) {
	switch a.Method {
	case "treatment":
		res, err := c.Treatment(a.Key, a.BucketingKey, a.Feature, a.Attributes)
		return res.Treatment, err
	case "treatments", "treatmentWithConfig", "treatmentsWithConfig", "track":
		return "", fmt.Errorf("method '%s' is not yet implemented", a.Method)
	default:
		return "", fmt.Errorf("unknwon method '%s'", a.Method)
	}
}
