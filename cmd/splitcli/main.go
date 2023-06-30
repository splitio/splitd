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
	"github.com/splitio/splitd/splitio/link/client"
	"github.com/splitio/splitd/splitio/util"
)

func main() {

	args, err := parseArgs()
	if err != nil {
		fmt.Println("error parsing arguments: ", err.Error())
		os.Exit(1)
	}

	logger := logging.NewLogger(nil)

	c, err := link.Consumer(logger, args.linkOpts()...)
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
	fmt.Printf("took: %d\n", time.Since(before).Microseconds())
	if err != nil {
		logger.Error("error executing call: ", err.Error())
		os.Exit(3)
	}

	fmt.Println(result)
}

func executeCall(c client.Interface, a *cliArgs) (string, error) {
	switch a.method {
	case "treatment":
		return c.Treatment(a.key, a.bucketingKey, a.feature, a.attributes)
	case "treatments", "treatmentWithConfig", "treatmentsWithConfig", "track":
		return "", fmt.Errorf("method '%s' is not yet implemented", a.method)
	default:
		return "", fmt.Errorf("unknwon method '%s'", a.method)
	}
}

type cliArgs struct {
	connType     string
	connAddr     string
	bufSize      int
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

func (a *cliArgs) linkOpts() []link.Option {
	var ret []link.Option
	if a.connType != "" {
		ret = append(ret, link.WithSockType(a.connType))
	}
	if a.connAddr != "" {
		ret = append(ret, link.WithAddress(a.connAddr))
	}
	if a.bufSize != 0 {
		ret = append(ret, link.WithBufSize(a.bufSize))
	}
	return ret
}

func parseArgs() (*cliArgs, error) {
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
