package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/splitio/splitd/splitio/link"
	cc "github.com/splitio/splitd/splitio/util/conf"
)

type CliArgs struct {
	LogLevel       string
	Protocol       string
	Serialization  string
	ConnType       string
	ConnAddr       string
	BufSize        int
	ReadTimeoutMS  int
	WriteTimeoutMS int

	// command
	Method       string
	Key          string
	BucketingKey string
	Feature      string
	Features     []string
	TrafficType  string
	EventType    string
	EventVal     float64
	Attributes   map[string]interface{}
}

func (a *CliArgs) LinkOpts() (*link.ConsumerOptions, error) {

	opts := link.DefaultConsumerOptions()

	var err error
	if a.Protocol != "" {
		if opts.Consumer.Protocol, err = cc.ParseProtocolVersion(a.Protocol); err != nil {
			return nil, fmt.Errorf("invalid protocol version %s", a.Protocol)
		}
	}

	if a.ConnType != "" {
		if opts.Transfer.ConnType, err = cc.ParseConnType(a.ConnType); err != nil {
			return nil, fmt.Errorf("invalid connection type %s", a.ConnType)
		}
	}

	if a.Serialization != "" {
		if opts.Serialization, err = cc.ParseSerializer(a.Serialization); err != nil {
			return nil, fmt.Errorf("invalid serialization %s", a.Serialization)
		}
	}

	durationFromMS := func(i int) time.Duration { return time.Duration(i) * time.Millisecond }
	cc.SetIfNotEmpty(&opts.Transfer.Address, &a.ConnAddr)
	cc.SetIfNotEmpty(&opts.Transfer.BufferSize, &a.BufSize)
	cc.MapIfNotEmpty(&opts.Transfer.ReadTimeout, &a.ReadTimeoutMS, durationFromMS)
	cc.MapIfNotEmpty(&opts.Transfer.WriteTimeout, &a.WriteTimeoutMS, durationFromMS)
	return &opts, nil
}

func ParseCliArgs() (*CliArgs, error) {

	cliFlags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	p := cliFlags.String("protocol", "", "Protocol version [v1]")
	s := cliFlags.String("serialization", "", "Client-Daemon communication serialization mechanism [msgpack]")
	ll := cliFlags.String("log-level", "INFO", "log level [ERROR,WARNING,INFO,DEBUG]")
	ct := cliFlags.String("conn-type", "", "unix-seqpacket|unix-stream")
	ca := cliFlags.String("conn-address", "", "path/ipv4-address")
	bs := cliFlags.Int("buffer-size", 0, "read buffer size in bytes")
	m := cliFlags.String("method", "", "treatment|treatments|treatmentWithConfig|treatmentsWithConfig|track")
	k := cliFlags.String("key", "", "user key")
	bk := cliFlags.String("bucketing-key", "", "bucketing key")
	f := cliFlags.String("feature", "", "feature to evaluate")
	fs := cliFlags.String("features", "", "features to evaluate (comma-separated list with no spaces in between)")
	tt := cliFlags.String("traffic-type", "", "traffic type of event")
	et := cliFlags.String("event-type", "", "event type")
	ev := cliFlags.String("value", "", "event associated value")
	at := cliFlags.String("attributes", "", "json representation of attributes")
	err := cliFlags.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("error parsing arguments: %w", err)
	}

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

	return &CliArgs{
        Serialization: *s,
		Protocol:     *p,
		LogLevel:     *ll,
		ConnType:     *ct,
		ConnAddr:     *ca,
		BufSize:      *bs,
		Method:       *m,
		Key:          *k,
		BucketingKey: *bk,
		Feature:      *f,
		Features:     strings.Split(*fs, ","),
		TrafficType:  *tt,
		EventType:    *et,
		EventVal:     val,
		Attributes:   attrs,
	}, nil
}
