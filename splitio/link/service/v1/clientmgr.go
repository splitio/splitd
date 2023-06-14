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
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("CRITICAL - connection handlers are panicking: ", r)
		}
	}()
	err := m.handleClientInteractions()
	if err != nil {
		m.logger.Error(fmt.Sprintf("an error occured when interacting with the client: %s. aborting", err))
	}
}

func (m *ClientManager) handleClientInteractions() error {
	defer m.cc.Shutdown()
	for {
		rpc, err := m.fetchRPC()
		if err != nil {
			if errors.Is(err, io.EOF) { // connection ended, no error
				m.logger.Debug(fmt.Sprintf("connection remotely closed for metadata=%+v", m.metadata))
				return nil
			} else if errors.Is(err, os.ErrDeadlineExceeded) { // we waited for an RPC, got none, try again.
				m.logger.Debug(fmt.Sprintf("read timeout/no RPC fetched. restarting loop for metadata=%+v", m.metadata))
				continue
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
	if err = m.serializer.Parse(read, &parsed); err != nil {
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
		if err := args.PopulateFromRPC(rpc); err != nil {
			return nil, fmt.Errorf("error parsing register arguments: %w", err)
		}
		return m.handleRegistration(&args)
	case protov1.OCTreatment:
		var args protov1.TreatmentArgs
		if err := args.PopulateFromRPC(rpc); err != nil {
			return nil, fmt.Errorf("error parsing treatment arguments: %w", err)
		}
		return m.handleGetTreatment(&args)
	case protov1.OCTreatments:
		var args protov1.TreatmentsArgs
		if err := args.PopulateFromRPC(rpc); err != nil {
			return nil, fmt.Errorf("error parsing treatments arguments: %w", err)
		}
		return m.handleGetTreatments(&args)
	}
	return nil, fmt.Errorf("RPC not implemented")
}

func (m *ClientManager) handleRegistration(args *protov1.RegisterArgs) (interface{}, error) {
	m.metadata = &types.ClientMetadata{
		ID:                   args.ID,
		SdkVersion:           args.SDKVersion,
		ReturnImpressionData: (args.Flags & protov1.RegisterFlagReturnImpressionData) != 0,
	}
	return &protov1.ResponseWrapper[protov1.RegisterPayload]{Status: protov1.ResultOk}, nil
}

func (m *ClientManager) handleGetTreatment(args *protov1.TreatmentArgs) (interface{}, error) {
	res, err := m.splitSDK.Treatment(m.metadata, args.Key, args.BucketingKey, args.Feature, args.Attributes)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}, err
	}

	response := &protov1.ResponseWrapper[protov1.TreatmentPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TreatmentPayload{Treatment: res.Treatment},
	}

	if m.metadata.ReturnImpressionData && res.Impression != nil {
		response.Payload.ListenerData = &protov1.ListenerExtraData{
			Label:        res.Impression.Label,
			Timestamp:    res.Impression.Time,
			ChangeNumber: res.Impression.ChangeNumber,
		}
	}

	return response, nil
}

func (m *ClientManager) handleGetTreatments(args *protov1.TreatmentsArgs) (interface{}, error) {
	res, err := m.splitSDK.Treatments(m.metadata, args.Key, args.BucketingKey, args.Features, args.Attributes)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}, err
	}

	results := make([]protov1.TreatmentPayload, len(args.Features))
	for idx, feature := range args.Features {
		ff, ok := res[feature]
		if !ok {
			results[idx].Treatment = "control"
			continue
		}

		results[idx].Treatment = ff.Treatment
		if ff.Impression != nil {
			results[idx].ListenerData = &protov1.ListenerExtraData{
				Label:        ff.Impression.Label,
				Timestamp:    ff.Impression.Time,
				ChangeNumber: ff.Impression.ChangeNumber,
			}
		}
	}

	response := &protov1.ResponseWrapper[protov1.TreatmentsPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TreatmentsPayload{Results: results},
	}

	return response, nil
}
