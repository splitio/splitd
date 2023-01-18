package v1

import (
	"errors"
	"fmt"
	"io"

	"github.com/splitio/go-toolkit/v5/logging"

	"github.com/splitio/splitd/splitio/link/listeners"
	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/sdk"
)

type ClientManager struct {
	cc         listeners.ClientConnection
	serializer serializer.Interface[protov1.RPC, protov1.Response]
	logger     logging.LoggerInterface
	metadata   *sdk.ClientMetadata
	splitSDK   sdk.Interface
	sdkFacade  sdk.Interface
}

func NewClientManager(
	cc listeners.ClientConnection,
	logger logging.LoggerInterface,
	splitSDK sdk.Interface,
	serializer serializer.Interface[protov1.RPC, protov1.Response],
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

	metadata, err := m.awaitRegistration()
	if err != nil {
		m.logger.Error(fmt.Sprintf("client registration failed: %s. aborting", err))
		return
	}

	m.metadata = metadata
	err = m.handleClientInteractions()
	if err != nil {
		m.logger.Error(fmt.Sprintf("an error occured when interacting with the client: %s. aboting", err))
		return
	}

}

func (m *ClientManager) awaitRegistration() (*sdk.ClientMetadata, error) {
	rpc, err := m.fetchRPC()
	if err != nil {
		return nil, fmt.Errorf("error reading initial register rpc: %w", err)
	}

	var args protov1.RegisterArgs
	err = args.PopulateFromRPC(rpc)
	if err != nil {
		return nil, fmt.Errorf("error parsing arguments: %w", err)
	}

	response := protov1.ResponseWrapper[protov1.RegisterPayload]{
		Status: protov1.ResultOk,
	}

	if err = m.sendResponse(&response); err != nil {
		return nil, err
	}

	return &sdk.ClientMetadata{
		ID:         args.ID,
		SdkVersion: args.SDKVersion,
	}, nil
}

func (m *ClientManager) handleClientInteractions() error {
	for {
		rpc, err := m.fetchRPC()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			m.logger.Error(fmt.Sprintf("error reading RPC: %s", err))
			continue
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

	parsed, err := m.serializer.Parse(read)
	if err != nil {
		return nil, fmt.Errorf("error parsing message: %w", err)
	}
	return parsed, nil
}

func (m *ClientManager) sendResponse(response protov1.Response) error {
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

func (m *ClientManager) handleRPC(rpc *protov1.RPC) (protov1.Response, error) {
	switch rpc.OpCode {
	case protov1.OCRegister:
		// TODO(mredolatti):
	case protov1.OCTreatment:
		var args protov1.TreatmentArgs
		err := args.PopulateFromRPC(rpc)
		if err != nil {
			return nil, fmt.Errorf("error parsing arguments: %w", err)
		}
		return m.handleGetTreatment(&args), nil
	}
	panic("not implemented")
}

func (m *ClientManager) handleGetTreatment(args *protov1.TreatmentArgs) protov1.Response {
	treatment, err := m.splitSDK.Treatment(m.metadata, args.Key, args.BucketingKey, args.Feature, args.Attributes)
	if err != nil {
		// TODO(mredolatti): Log!
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}
	}
	return &protov1.ResponseWrapper[protov1.TreatmentPayload]{
		Status: protov1.ResultOk,
		Payload: protov1.TreatmentPayload{
			Treatment: treatment,
		},
	}
}
