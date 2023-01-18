package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/clients"
	"github.com/splitio/splitd/splitio/consumer"
	"github.com/splitio/splitd/splitio/util"
)

func main() {

	cfg := parseConfig()

	client, err := clients.Create(
		clients.ConnTypeUnixSeqPacket, // TODO(mredolatti): parse from cfg/cli-args
		clients.WithFileName(cfg.sockAddr),
	)
	mustNotFail(err)

	shutdown := util.NewShutdownHandler()
	shutdown.RegisterHook(func() {
		err := client.Shutdown()
		if err != nil {
			fmt.Println(err.Error())
		}
	})
	defer shutdown.TriggerAndWait()

	c := consumer.New(logging.NewExtendedLogger(nil), client)

	before := time.Now()
	treatment, err := c.Treatment("key1", "bk1", "feat1", map[string]interface{}{
		"a": 1,
		"b": "asd",
		"c": []string{"q", "w", "e"},
	})
	mustNotFail(err)
	after := time.Since(before)

	fmt.Println(treatment)
	fmt.Printf("took: %dns\n", after.Nanoseconds())

	

}


type config struct {
	sockType string
	sockAddr string
}

func parseConfig() *config {
	st := flag.String("socket-type", "unix-seqpacket", "unix-seqpacket|unix-stream")
	sa := flag.String("socket-address", "/var/run/splitd.sock", "path/ipv4-address of the socket")
	flag.Parse()
	return &config{
		sockType: *st,
		sockAddr: *sa,
	}

}

func mustNotFail(err error) {
	if err != nil {
		panic(err.Error())
	}
}
