package client

import (
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/clients"
	"github.com/splitio/splitd/splitio/protocol"
	"github.com/vmihailenco/msgpack/v5"
)

type Interface interface {
	Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (string, error)
}

type Impl struct {
	logger logging.LoggerInterface
	client clients.Raw
}

func New(logger logging.LoggerInterface, client clients.Raw) *Impl {
	return &Impl{
		logger: logger,
		client: client,
	}
}

// Treatment implements Interface
func (c *Impl) Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (string, error) {
	rpc := protocol.RPC{
		Version: protocol.V1,
		OpCode:  protocol.OCTreatment,
		Args:    []interface{}{key, bucketingKey, feature, attrs},
	}

	serialized, err := msgpack.Marshal(&rpc)
	if err != nil {
		return "control", fmt.Errorf("error serializing rpc: %w", err)
	}

	err = c.client.SendMessage(serialized)
	if err != nil {
		return "control", fmt.Errorf("error sending message to split daemon: %w", err)
	}

	resp, err := c.client.ReceiveMessage()
	if err != nil {
		return "control", fmt.Errorf("error reading response from daeom: %w", err)
	}

	return string(resp), nil
}

var _ Interface = (*Impl)(nil)
