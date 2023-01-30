package v1

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"

	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/sdk/types"
)

type ClientManager struct {
	cc         transfer.RawConn
	serializer serializer.Interface
	logger     logging.LoggerInterface
	metadata   *types.ClientMetadata
	splitSDK   sdk.Interface
	sdkFacade  sdk.Interface
}

func NewClientManager(
	cc transfer.RawConn,
	logger logging.LoggerInterface,
	splitSDK sdk.Interface,
	serializer serializer.Interface,
) *ClientManager {
	return &ClientManager{
		cc:         cc,
		logger:     logger,
		serializer: serializer,
		splitSDK:   splitSDK,
	}
}

func (m *ClientManager) Manage() {
	defer m.cc.Shutdown()

	err := m.handleClientInteractions()
	if err != nil {
		m.logger.Error(fmt.Sprintf("an error occured when interacting with the client: %s. aboting", err))
		return
	}
}

func (m *ClientManager) handleClientInteractions() error {
	for {
		rpc, err := m.fetchRPC()
		if err != nil {
			if errors.Is(err, io.EOF) {
				m.logger.Debug(fmt.Sprintf("connection remotely closed for metadata=%+v", m.metadata))
				return nil // connection ended, no error
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				m.logger.Debug(fmt.Sprintf("read timeout/no RPC fetched. restarting loop for metadata=%+v", m.metadata))
				continue // we waited for an RPC, got none, try again.
			} else {
				m.logger.Error(fmt.Sprintf("unexpected error reading RPC: %s. Closing conn for metadata=%+v", err, m.metadata))
				return err
			}
		}

		response, err := m.handleRPC(rpc)
		if err != nil {
			// TODO(mredolatti): see if this is recoverable
			return fmt.Errorf("error handling RPC: %w", err)
		}

		if err = m.sendResponse(response); err != nil {
			return err
		}
	}
}

func (m *ClientManager) fetchRPC() (*protov1.RPC, error) {
	read, err := m.cc.ReceiveMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading from conn: %w", err)
	}

	var parsed protov1.RPC
	err = m.serializer.Parse(read, &parsed)
	if err != nil {
		return nil, fmt.Errorf("error parsing message: %w", err)
	}
	return &parsed, nil
}

func (m *ClientManager) sendResponse(response interface{}) error {
	serialized, err := m.serializer.Serialize(response)
	if err != nil {
		// TODO(mredolatti): see if this is recoverable
		return fmt.Errorf("error serializing response: %w", err)
	}

	err = m.cc.SendMessage(serialized)
	if err != nil {
		// TODO(mredolatti): see if this is recoverable
		return fmt.Errorf("error sending response back to the client: %w", err)
	}

	return nil
}

func (m *ClientManager) handleRPC(rpc *protov1.RPC) (interface{}, error) {

	if m.metadata == nil && rpc.OpCode != protov1.OCRegister {
		return nil, fmt.Errorf("first call must be 'register'`")
	}

	switch rpc.OpCode {
	case protov1.OCRegister:
		var args protov1.RegisterArgs
		err := args.PopulateFromRPC(rpc)
		if err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}
		return m.handleRegistration(&args)
	case protov1.OCTreatment:
		var args protov1.TreatmentArgs
		err := args.PopulateFromRPC(rpc)
		if err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}
		return m.handleGetTreatment(&args)
	}
	return nil, fmt.Errorf("RPC not implemented")
}

func (m *ClientManager) handleRegistration(args *protov1.RegisterArgs) (interface{}, error) {
	m.metadata = &types.ClientMetadata{
		ID:         args.ID,
		SdkVersion: args.SDKVersion,
	}
	return protov1.ResponseWrapper[protov1.RegisterPayload]{Status: protov1.ResultOk}, nil

}

func (m *ClientManager) handleGetTreatment(args *protov1.TreatmentArgs) (interface{}, error) {
	treatment, err := m.splitSDK.Treatment(m.metadata, args.Key, args.BucketingKey, args.Feature, args.Attributes)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}, err
	}
	return &protov1.ResponseWrapper[protov1.TreatmentPayload]{
		Status: protov1.ResultOk,
		Payload: protov1.TreatmentPayload{
			Treatment: treatment,
		},
	}, nil
}
