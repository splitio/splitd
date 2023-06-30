package client

import (
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	clientv1 "github.com/splitio/splitd/splitio/link/client/v1"
	"github.com/splitio/splitd/splitio/link/common"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
)

type Interface interface {
	Treatment(key string, bk string, feature string, attributes map[string]interface{}) (string, error)
	Shutdown() error
}

func New(logger logging.LoggerInterface, conn transfer.RawConn, os ...common.Option) (Interface, error) {
	o := common.DefaultOpts()
	o.Parse(os)

	s, err := serializer.Setup(o.Serial)
	if err != nil {
		return nil, fmt.Errorf("error building serializer")
	}

	switch o.ProtoV {
	case protocol.V1:
		return clientv1.New(logger, conn, s)
	}
	return nil, fmt.Errorf("unknown protocol version: '%d'", o.ProtoV)
}
