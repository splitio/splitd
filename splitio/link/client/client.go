package client

import (
	"fmt"
	"os"
	"strconv"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/client/types"
	clientv1 "github.com/splitio/splitd/splitio/link/client/v1"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
)

func New(logger logging.LoggerInterface, conn transfer.RawConn, serial serializer.Interface, opts Options) (types.ClientInterface, error) {
	switch opts.Protocol {
	case protocol.V1:
		return clientv1.New(opts.ID, logger, conn, serial, opts.ImpressionsFeedback)
	}
	return nil, fmt.Errorf("unknown protocol version: '%d'", opts.Protocol)
}

type Options struct {
	ID                  string
	Protocol            protocol.Version
	ImpressionsFeedback bool
}

func DefaultOptions() Options {
	return Options{
		ID:                  strconv.Itoa(os.Getpid()),
		Protocol:            protocol.V1,
		ImpressionsFeedback: false,
	}
}
