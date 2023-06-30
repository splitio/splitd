package v1

import (
	"fmt"
	"os"
	"strconv"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
	"github.com/splitio/splitd/splitio/link/protocol"
	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
)

type Interface interface {
	Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (string, error)
}

type Impl struct {
	logger     logging.LoggerInterface
	conn       transfer.RawConn
	serializer serializer.Interface
}

func New(logger logging.LoggerInterface, conn transfer.RawConn, serializer serializer.Interface) (*Impl, error) {
	i := &Impl{
		logger:     logger,
		conn:       conn,
		serializer: serializer,
	}

	if err := i.register(); err != nil {
		i.conn.Shutdown()
		return nil, fmt.Errorf("error during client registration: %w", err)
	}

	return i, nil
}

// Treatment implements Interface
func (c *Impl) Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (string, error) {
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCTreatment,
		Args:    []interface{}{key, bucketingKey, feature, attrs},
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.TreatmentPayload]](c, &rpc)
	if err != nil {
		return "control", fmt.Errorf("error executing rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return "control", fmt.Errorf("server responded with error %d", resp.Status)
	}

	return resp.Payload.Treatment, nil
}

func (c *Impl) register() error {
	rpc := protov1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  protov1.OCRegister,
		Args:    []interface{}{strconv.Itoa(os.Getpid()), fmt.Sprintf("splitd-%s", splitio.Version)},
	}

	resp, err := doRPC[protov1.ResponseWrapper[protov1.RegisterPayload]](c, &rpc)
	if err != nil {
		return fmt.Errorf("error executing rpc: %w", err)
	}

	if resp.Status != protov1.ResultOk {
		return fmt.Errorf("server responded with error %d", resp.Status)
	}

	return nil
}

func doRPC[T any](c *Impl, rpc *protov1.RPC) (*T, error) {
	serialized, err := c.serializer.Serialize(&rpc)
	if err != nil {
		return nil, fmt.Errorf("error serializing rpc: %w", err)
	}

	err = c.conn.SendMessage(serialized)
	if err != nil {
		return nil, fmt.Errorf("error sending message to split daemon: %w", err)
	}

	resp, err := c.conn.ReceiveMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading response from daeom: %w", err)
	}

	var response T
	err = c.serializer.Parse(resp, &response)
	if err != nil {
		return nil, fmt.Errorf("error de-serializing server response: %w", err)
	}

	return &response, nil
}

func (c *Impl) Shutdown() error {
	return c.conn.Shutdown()
}

var _ Interface = (*Impl)(nil)
