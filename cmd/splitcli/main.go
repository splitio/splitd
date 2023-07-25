package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link"
	"github.com/splitio/splitd/splitio/link/client/types"
	"github.com/splitio/splitd/splitio/util"
	cc "github.com/splitio/splitd/splitio/util/conf"
)

func main() {

	args, err := parseArgs()
	if err != nil {
		fmt.Println("error parsing arguments: ", err.Error())
		os.Exit(1)
	}

	linkOpts, err := args.linkOpts()
	if err != nil {
		fmt.Println("error building options from arguments: ", err.Error())
		os.Exit(1)
	}

	logLevel := logging.Level(args.logLevel)
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

func executeCall(c types.ClientInterface, a *cliArgs) (string, error) {
	switch a.method {
	case "treatment":
		res, err := c.Treatment(a.key, a.bucketingKey, a.feature, a.attributes)
		return res.Treatment, err
	case "treatments", "treatmentWithConfig", "treatmentsWithConfig", "track":
		return "", fmt.Errorf("method '%s' is not yet implemented", a.method)
	default:
		return "", fmt.Errorf("unknwon method '%s'", a.method)
	}
}

type cliArgs struct {
	logLevel       string
	protocol       string
	serialization  string
	connType       string
	connAddr       string
	bufSize        int
	readTimeoutMS  int
	writeTimeoutMS int

	// command
	method       string
	key          string
	bucketingKey string
	feature      string
	features     []string
	trafficType  string
	eventType    string
	eventVal     float64
	attributes   map[string]interface{}
}

func (a *cliArgs) linkOpts() (*link.ConsumerOptions, error) {

	opts := link.DefaultConsumerOptions()

	var err error
	if a.protocol != "" {
		if opts.Consumer.Protocol, err = cc.ParseProtocolVersion(a.protocol); err != nil {
			return nil, fmt.Errorf("invalid protocol version %s", a.protocol)
		}
	}

	if a.connType != "" {
		if opts.Transfer.ConnType, err = cc.ParseConnType(a.connType); err != nil {
			return nil, fmt.Errorf("invalid connection type %s", a.connType)
		}
	}

	if a.serialization != "" {
		if opts.Serialization, err = cc.ParseSerializer(a.serialization); err != nil {
			return nil, fmt.Errorf("invalid serialization %s", a.serialization)
		}
	}

	durationFromMS := func(i int) time.Duration { return time.Duration(i) * time.Millisecond }
	cc.SetIfNotEmpty(&opts.Transfer.Address, &a.connAddr)
	cc.SetIfNotEmpty(&opts.Transfer.BufferSize, &a.bufSize)
	cc.MapIfNotEmpty(&opts.Transfer.ReadTimeout, &a.readTimeoutMS, durationFromMS)
	cc.MapIfNotEmpty(&opts.Transfer.WriteTimeout, &a.writeTimeoutMS, durationFromMS)

	return &opts, nil
}

func parseArgs() (*cliArgs, error) {
	ll := flag.String("log-level", "INFO", "log level [ERROR,WARNING,INFO,DEBUG]")
	ct := flag.String("conn-type", "", "unix-seqpacket|unix-stream")
	ca := flag.String("conn-address", "", "path/ipv4-address")
	bs := flag.Int("buffer-size", 0, "read buffer size in bytes")
	m := flag.String("method", "", "treatment|treatments|treatmentWithConfig|treatmentsWithConfig|track")
	k := flag.String("key", "", "user key")
	bk := flag.String("bucketing-key", "", "bucketing key")
	f := flag.String("feature", "", "feature to evaluate")
	fs := flag.String("features", "", "features to evaluate (comma-separated list with no spaces in between)")
	tt := flag.String("traffic-type", "", "traffic type of event")
	et := flag.String("event-type", "", "event type")
	ev := flag.String("value", "", "event associated value")
	at := flag.String("attributes", "", "json representation of attributes")

	flag.Parse()

	val, err := strconv.ParseFloat(*ev, 64)
	if *ev != "" && err != nil {
		return nil, fmt.Errorf("error parsing event value")
	}

	if *at == "" {
		*at = "null"
	}
	attrs := make(map[string]interface{})
	if err = json.Unmarshal([]byte(*at), &attrs); err != nil {
		return nil, fmt.Errorf("error parsing attributes: %w", err)
	}

	return &cliArgs{
		logLevel:     *ll,
		connType:     *ct,
		connAddr:     *ca,
		bufSize:      *bs,
		method:       *m,
		key:          *k,
		bucketingKey: *bk,
		feature:      *f,
		features:     strings.Split(*fs, ","),
		trafficType:  *tt,
		eventType:    *et,
		eventVal:     val,
		attributes:   attrs,
	}, nil

}
