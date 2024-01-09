package v1

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/splitio/go-toolkit/v5/logging"

	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/sdk/types"
)

type ClientManager struct {
	cc           transfer.RawConn
	serializer   serializer.Interface
	logger       logging.LoggerInterface
	clientConfig *types.ClientConfig
	splitSDK     sdk.Interface
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
			m.logger.Error("CRITICAL - connection handler is panicking: ", r)
			m.logger.Error(string(debug.Stack())) // debug.Stack() returns the panic's stack when called in a recover block
		}
	}()
	err := m.handleClientInteractions()
	if err != nil {
		m.logger.Error(fmt.Sprintf("an error occured when interacting with the client: %s. aborting", err))
	}
}

func (m *ClientManager) handleClientInteractions() error {
	for {
		rpc, err := m.fetchRPC()
		if err != nil {
			if errors.Is(err, io.EOF) { // connection ended, no error
				m.logger.Debug(fmt.Sprintf("connection remotely closed for metadata=%s", formatClientConfig(m.clientConfig)))
				return nil
			} else if errors.Is(err, os.ErrDeadlineExceeded) { // we waited for an RPC, got none, try again.
				m.logger.Debug(fmt.Sprintf("read timeout/no RPC fetched. restarting loop for metadata=%s", formatClientConfig(m.clientConfig)))
				continue
			} else {
				m.logger.Error(fmt.Sprintf("unexpected error reading RPC: %s. Closing conn for metadata=%s", err, formatClientConfig(m.clientConfig)))
				return err
			}
		}

		response, err := m.dispatchRPC(rpc)
		if err != nil {
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
		return fmt.Errorf("error serializing response: %w", err)
	}

	err = m.cc.SendMessage(serialized)
	if err != nil {
		return fmt.Errorf("error sending response back to the client: %w", err)
	}

	return nil
}

func (m *ClientManager) dispatchRPC(rpc *protov1.RPC) (interface{}, error) {

	if m.clientConfig == nil && rpc.OpCode != protov1.OCRegister {
		return nil, fmt.Errorf("first call must be 'register'`")
	}

	switch rpc.OpCode {
	case protov1.OCRegister:
		return m.handleRegistration(rpc)
	case protov1.OCTreatment:
		return m.handleGetTreatment(rpc, false)
	case protov1.OCTreatments:
		return m.handleGetTreatments(rpc, false)
	case protov1.OCTreatmentWithConfig:
		return m.handleGetTreatment(rpc, true)
	case protov1.OCTreatmentsWithConfig:
		return m.handleGetTreatments(rpc, true)
	case protov1.OCTreatmentsByFlagSet:
		return m.handleGetTreatmentsByFlagSet(rpc, false)
	case protov1.OCTreatmentsWithConfigByFlagSet:
		return m.handleGetTreatmentsByFlagSet(rpc, true)
	case protov1.OCTreatmentsByFlagSets:
		return m.handleGetTreatmentsByFlagSets(rpc, false)
	case protov1.OCTreatmentsWithConfigByFlagSets:
		return m.handleGetTreatmentsByFlagSets(rpc, true)
	case protov1.OCTrack:
		return m.handleTrack(rpc)
	case protov1.OCSplitNames:
		return m.handleSplitNames(rpc)
	case protov1.OCSplit:
		return m.handleSplit(rpc)
	case protov1.OCSplits:
		return m.handleSplits(rpc)
	}

	return nil, fmt.Errorf("RPC not implemented")
}

func (m *ClientManager) handleRegistration(rpc *protov1.RPC) (interface{}, error) {

	var args protov1.RegisterArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing register arguments: %w", err)
	}

	m.clientConfig = &types.ClientConfig{
		Metadata: types.ClientMetadata{
			ID:         args.ID,
			SdkVersion: args.SDKVersion,
		},
		ReturnImpressionData: (args.Flags & protov1.RegisterFlagReturnImpressionData) != 0,
	}
	return &protov1.ResponseWrapper[protov1.RegisterPayload]{Status: protov1.ResultOk}, nil
}

func (m *ClientManager) handleGetTreatment(rpc *protov1.RPC, withConfig bool) (interface{}, error) {

	var args protov1.TreatmentArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing treatment arguments: %w", err)
	}

	res, err := m.splitSDK.Treatment(m.clientConfig, args.Key, args.BucketingKey, args.Feature, args.Attributes)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}, err
	}

	response := &protov1.ResponseWrapper[protov1.TreatmentPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TreatmentPayload{Treatment: res.Treatment},
	}

	if withConfig {
		response.Payload.Config = res.Config
	}

	if m.clientConfig.ReturnImpressionData && res.Impression != nil {
		response.Payload.ListenerData = &protov1.ListenerExtraData{
			Label:        res.Impression.Label,
			Timestamp:    res.Impression.Time,
			ChangeNumber: res.Impression.ChangeNumber,
		}
	}

	return response, nil
}

func (m *ClientManager) handleGetTreatments(rpc *protov1.RPC, withConfig bool) (interface{}, error) {

	var args protov1.TreatmentsArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing treatments arguments: %w", err)
	}

	res, err := m.splitSDK.Treatments(m.clientConfig, args.Key, args.BucketingKey, args.Features, args.Attributes)
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
		if m.clientConfig.ReturnImpressionData && ff.Impression != nil {
			results[idx].ListenerData = &protov1.ListenerExtraData{
				Label:        ff.Impression.Label,
				Timestamp:    ff.Impression.Time,
				ChangeNumber: ff.Impression.ChangeNumber,
			}
		}

		if withConfig {
			results[idx].Config = ff.Config
		}
	}

	response := &protov1.ResponseWrapper[protov1.TreatmentsPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TreatmentsPayload{Results: results},
	}

	return response, nil
}

func (m *ClientManager) handleGetTreatmentsByFlagSet(rpc *protov1.RPC, withConfig bool) (interface{}, error) {

	var args protov1.TreatmentsByFlagSetArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing treatmentsByFlagSet arguments: %w", err)
	}

	res, err := m.splitSDK.TreatmentsByFlagSet(m.clientConfig, args.Key, args.BucketingKey, args.FlagSet, args.Attributes)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{Status: protov1.ResultInternalError}, err
	}

	results := make(map[string]protov1.TreatmentPayload, len(res))
	for feature, evaluationResult := range res {
		currentPayload := protov1.TreatmentPayload{
			Treatment: evaluationResult.Treatment,
		}

		if m.clientConfig.ReturnImpressionData && evaluationResult.Impression != nil {
			currentPayload.ListenerData = &protov1.ListenerExtraData{
				Label:        evaluationResult.Impression.Label,
				Timestamp:    evaluationResult.Impression.Time,
				ChangeNumber: evaluationResult.Impression.ChangeNumber,
			}
		}

		if withConfig {
			currentPayload.Config = evaluationResult.Config
		}

		results[feature] = currentPayload
	}

	response := &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TreatmentsWithFeaturePayload{Results: results},
	}

	return response, nil
}

func (m *ClientManager) handleGetTreatmentsByFlagSets(rpc *protov1.RPC, withConfig bool) (interface{}, error) {

	var args protov1.TreatmentsByFlagSetsArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing treatmentsByFlagSets arguments: %w", err)
	}

	res, err := m.splitSDK.TreatmentsByFlagSets(m.clientConfig, args.Key, args.BucketingKey, args.FlagSets, args.Attributes)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{Status: protov1.ResultInternalError}, err
	}

	results := make(map[string]protov1.TreatmentPayload, len(res))
	for feature, evaluationResult := range res {
		currentPayload := protov1.TreatmentPayload{
			Treatment: evaluationResult.Treatment,
		}

		if m.clientConfig.ReturnImpressionData && evaluationResult.Impression != nil {
			currentPayload.ListenerData = &protov1.ListenerExtraData{
				Label:        evaluationResult.Impression.Label,
				Timestamp:    evaluationResult.Impression.Time,
				ChangeNumber: evaluationResult.Impression.ChangeNumber,
			}
		}

		if withConfig {
			currentPayload.Config = evaluationResult.Config
		}

		results[feature] = currentPayload
	}

	response := &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TreatmentsWithFeaturePayload{Results: results},
	}

	return response, nil
}

func (m *ClientManager) handleTrack(rpc *protov1.RPC) (interface{}, error) {

	var args protov1.TrackArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing track arguments: %w", err)
	}

	err := m.splitSDK.Track(m.clientConfig, args.Key, args.TrafficType, args.EventType, args.Value, args.Properties)
	if err != nil && !errors.Is(err, sdk.ErrEventsQueueFull) {
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}, err
	}

	response := &protov1.ResponseWrapper[protov1.TrackPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.TrackPayload{Success: err == nil}, // if err != nil it can only be ErrEventsQueueFull at this point
	}

	return response, nil
}

func (m *ClientManager) handleSplitNames(rpc *protov1.RPC) (interface{}, error) {

	var args protov1.SplitNamesArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing split-names arguments: %w", err)
	}

	names, err := m.splitSDK.SplitNames()
	if err != nil {
		return &protov1.ResponseWrapper[protov1.SplitNamesPayload]{Status: protov1.ResultInternalError}, err
	}

	response := &protov1.ResponseWrapper[protov1.SplitNamesPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.SplitNamesPayload{Names: names},
	}

	return response, nil
}

func (m *ClientManager) handleSplit(rpc *protov1.RPC) (interface{}, error) {

	var args protov1.SplitArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing split arguments: %w", err)
	}

	view, err := m.splitSDK.Split(args.Name)
	if err != nil {
		return &protov1.ResponseWrapper[protov1.TreatmentPayload]{Status: protov1.ResultInternalError}, err
	}

	response := &protov1.ResponseWrapper[protov1.SplitPayload]{
		Status:  protov1.ResultOk,
		Payload: protov1.SplitPayload(*view),
	}

	return response, nil
}

func (m *ClientManager) handleSplits(rpc *protov1.RPC) (interface{}, error) {

	var args protov1.SplitsArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing splits arguments: %w", err)
	}

	views, err := m.splitSDK.Splits()
	if err != nil {
		return &protov1.ResponseWrapper[protov1.SplitsPayload]{Status: protov1.ResultInternalError}, err
	}

	var p protov1.SplitsPayload
	p.Splits = make([]protov1.SplitPayload, 0, len(views))
	for _, view := range views {
		p.Splits = append(p.Splits, protov1.SplitPayload(view))
	}

	response := &protov1.ResponseWrapper[protov1.SplitsPayload]{
		Status:  protov1.ResultOk,
		Payload: p,
	}

	return response, nil
}

func formatClientConfig(c *types.ClientConfig) string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%+v", c)
}
