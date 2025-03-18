package v1

import (
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/sdk/types"
)

type UnBufferedClientStateMachineImpl struct {
	serializer   serializer.Interface
	logger       logging.LoggerInterface
	clientConfig *types.ClientConfig
	splitSDK     sdk.Interface
}

// HandleIncomingData implements ClientStateMachine.
func (u *UnBufferedClientStateMachineImpl) HandleIncomingData(in []byte, out *[]byte) (int, error) {

	var rpc protov1.RPC
	if err := u.serializer.Parse(in, &rpc); err != nil {
		return 0, fmt.Errorf("error parsing message: %w", err)
	}

	res, err := u.dispatchRPC(&rpc)
	if err != nil {
		return 0, fmt.Errorf("error handling RPC: %w", err)
	}

	serialized, err := u.serializer.Serialize(res)
	if err != nil {
		return 0, fmt.Errorf("error serializing response: %w", err)
	}

	n := copy(*out, serialized)
	if n < len(serialized) { // TODO(mredolatti): Remove this!
		return 0, fmt.Errorf("buffer too small, needed %d bytes, have %d bytes", len(serialized), len(*out))
	}

	return n, nil
}

func (u *UnBufferedClientStateMachineImpl) dispatchRPC(rpc *protov1.RPC) (interface{}, error) {
	switch rpc.OpCode {
	case protov1.OCRegister:
		return u.handleRegistration(rpc)
	case protov1.OCTreatment:
		return u.handleGetTreatment(rpc, false)
	//case protov1.OCTreatments:
	//	return u.handleGetTreatments(rpc, false)
	case protov1.OCTreatmentWithConfig:
		return u.handleGetTreatment(rpc, true)
	//case protov1.OCTreatmentsWithConfig:
	//	return u.handleGetTreatments(rpc, true)
	//case protov1.OCTreatmentsByFlagSet:
	//	return u.handleGetTreatmentsByFlagSet(rpc, false)
	//case protov1.OCTreatmentsWithConfigByFlagSet:
	//	return u.handleGetTreatmentsByFlagSet(rpc, true)
	//case protov1.OCTreatmentsByFlagSets:
	//	return u.handleGetTreatmentsByFlagSets(rpc, false)
	//case protov1.OCTreatmentsWithConfigByFlagSets:
	//	return u.handleGetTreatmentsByFlagSets(rpc, true)
	//case protov1.OCTrack:
	//	return u.handleTrack(rpc)
	case protov1.OCSplitNames:
		return u.handleSplitNames(rpc)
		//case protov1.OCSplit:
		//	return u.handleSplit(rpc)
		//case protov1.OCSplits:
		//	return u.handleSplits(rpc)
	}
	return nil, fmt.Errorf("RPC not implemented")
}

func (u *UnBufferedClientStateMachineImpl) handleRegistration(rpc *protov1.RPC) (interface{}, error) {

	var args protov1.RegisterArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing register arguments: %w", err)
	}

	u.clientConfig = &types.ClientConfig{
		Metadata: types.ClientMetadata{
			ID:         args.ID,
			SdkVersion: args.SDKVersion,
		},
		ReturnImpressionData: (args.Flags & protov1.RegisterFlagReturnImpressionData) != 0,
	}
	return &protov1.ResponseWrapper[protov1.RegisterPayload]{Status: protov1.ResultOk}, nil
}

func (u *UnBufferedClientStateMachineImpl) handleGetTreatment(rpc *protov1.RPC, withConfig bool) (interface{}, error) {

	var args protov1.TreatmentArgs
	if err := args.PopulateFromRPC(rpc); err != nil {
		return nil, fmt.Errorf("error parsing treatment arguments: %w", err)
	}

	res, err := u.splitSDK.Treatment(u.clientConfig, args.Key, args.BucketingKey, args.Feature, args.Attributes)
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

	if u.clientConfig.ReturnImpressionData && res.Impression != nil {
		response.Payload.ListenerData = &protov1.ListenerExtraData{
			Label:        res.Impression.Label,
			Timestamp:    res.Impression.Time,
			ChangeNumber: res.Impression.ChangeNumber,
		}
	}

	return response, nil
}

// func (m *ClientManager) handleGetTreatments(rpc *protov1.RPC, withConfig bool) (interface{}, error) {
//
//		var args protov1.TreatmentsArgs
//		if err := args.PopulateFromRPC(rpc); err != nil {
//			return nil, fmt.Errorf("error parsing treatments arguments: %w", err)
//		}
//
//		res, err := m.splitSDK.Treatments(m.clientConfig, args.Key, args.BucketingKey, args.Features, args.Attributes)
//		if err != nil {
//			return &protov1.ResponseWrapper[protov1.TreatmentsPayload]{Status: protov1.ResultInternalError}, err
//		}
//
//		results := make([]protov1.TreatmentPayload, len(args.Features))
//		for idx, feature := range args.Features {
//			ff, ok := res[feature]
//			if !ok {
//				results[idx].Treatment = "control"
//				continue
//			}
//
//			results[idx].Treatment = ff.Treatment
//			if m.clientConfig.ReturnImpressionData && ff.Impression != nil {
//				results[idx].ListenerData = &protov1.ListenerExtraData{
//					Label:        ff.Impression.Label,
//					Timestamp:    ff.Impression.Time,
//					ChangeNumber: ff.Impression.ChangeNumber,
//				}
//			}
//
//			if withConfig {
//				results[idx].Config = ff.Config
//			}
//		}
//
//		response := &protov1.ResponseWrapper[protov1.TreatmentsPayload]{
//			Status:  protov1.ResultOk,
//			Payload: protov1.TreatmentsPayload{Results: results},
//		}
//
//		return response, nil
//	}
//
// func (m *ClientManager) handleGetTreatmentsByFlagSet(rpc *protov1.RPC, withConfig bool) (interface{}, error) {
//
//		var args protov1.TreatmentsByFlagSetArgs
//		if err := args.PopulateFromRPC(rpc); err != nil {
//			return nil, fmt.Errorf("error parsing treatmentsByFlagSet arguments: %w", err)
//		}
//
//		res, err := m.splitSDK.TreatmentsByFlagSet(m.clientConfig, args.Key, args.BucketingKey, args.FlagSet, args.Attributes)
//		if err != nil {
//			return &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{Status: protov1.ResultInternalError}, err
//		}
//
//		results := make(map[string]protov1.TreatmentPayload, len(res))
//		for feature, evaluationResult := range res {
//			currentPayload := protov1.TreatmentPayload{
//				Treatment: evaluationResult.Treatment,
//			}
//
//			if m.clientConfig.ReturnImpressionData && evaluationResult.Impression != nil {
//				currentPayload.ListenerData = &protov1.ListenerExtraData{
//					Label:        evaluationResult.Impression.Label,
//					Timestamp:    evaluationResult.Impression.Time,
//					ChangeNumber: evaluationResult.Impression.ChangeNumber,
//				}
//			}
//
//			if withConfig {
//				currentPayload.Config = evaluationResult.Config
//			}
//
//			results[feature] = currentPayload
//		}
//
//		response := &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{
//			Status:  protov1.ResultOk,
//			Payload: protov1.TreatmentsWithFeaturePayload{Results: results},
//		}
//
//		return response, nil
//	}
//
// func (m *ClientManager) handleGetTreatmentsByFlagSets(rpc *protov1.RPC, withConfig bool) (interface{}, error) {
//
//		var args protov1.TreatmentsByFlagSetsArgs
//		if err := args.PopulateFromRPC(rpc); err != nil {
//			return nil, fmt.Errorf("error parsing treatmentsByFlagSets arguments: %w", err)
//		}
//
//		res, err := m.splitSDK.TreatmentsByFlagSets(m.clientConfig, args.Key, args.BucketingKey, args.FlagSets, args.Attributes)
//		if err != nil {
//			return &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{Status: protov1.ResultInternalError}, err
//		}
//
//		results := make(map[string]protov1.TreatmentPayload, len(res))
//		for feature, evaluationResult := range res {
//			currentPayload := protov1.TreatmentPayload{
//				Treatment: evaluationResult.Treatment,
//			}
//
//			if m.clientConfig.ReturnImpressionData && evaluationResult.Impression != nil {
//				currentPayload.ListenerData = &protov1.ListenerExtraData{
//					Label:        evaluationResult.Impression.Label,
//					Timestamp:    evaluationResult.Impression.Time,
//					ChangeNumber: evaluationResult.Impression.ChangeNumber,
//				}
//			}
//
//			if withConfig {
//				currentPayload.Config = evaluationResult.Config
//			}
//
//			results[feature] = currentPayload
//		}
//
//		response := &protov1.ResponseWrapper[protov1.TreatmentsWithFeaturePayload]{
//			Status:  protov1.ResultOk,
//			Payload: protov1.TreatmentsWithFeaturePayload{Results: results},
//		}
//
//		return response, nil
//	}
//
// func (m *ClientManager) handleTrack(rpc *protov1.RPC) (interface{}, error) {
//
//		var args protov1.TrackArgs
//		if err := args.PopulateFromRPC(rpc); err != nil {
//			return nil, fmt.Errorf("error parsing track arguments: %w", err)
//		}
//
//		err := m.splitSDK.Track(m.clientConfig, args.Key, args.TrafficType, args.EventType, args.Value, args.Properties)
//		if err != nil && !errors.Is(err, sdk.ErrEventsQueueFull) {
//			return &protov1.ResponseWrapper[protov1.TrackPayload]{Status: protov1.ResultInternalError}, err
//		}
//
//		response := &protov1.ResponseWrapper[protov1.TrackPayload]{
//			Status:  protov1.ResultOk,
//			Payload: protov1.TrackPayload{Success: err == nil}, // if err != nil it can only be ErrEventsQueueFull at this point
//		}
//
//		return response, nil
//	}
func (m *UnBufferedClientStateMachineImpl) handleSplitNames(rpc *protov1.RPC) (interface{}, error) {

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

//
//func (m *ClientManager) handleSplit(rpc *protov1.RPC) (interface{}, error) {
//
//	var args protov1.SplitArgs
//	if err := args.PopulateFromRPC(rpc); err != nil {
//		return nil, fmt.Errorf("error parsing split arguments: %w", err)
//	}
//
//	view, err := m.splitSDK.Split(args.Name)
//	if err != nil {
//		if errors.Is(err, sdk.ErrSplitNotFound) {
//			return &protov1.ResponseWrapper[protov1.SplitPayload]{Status: protov1.ResultOk}, nil
//		}
//		return &protov1.ResponseWrapper[protov1.SplitPayload]{Status: protov1.ResultInternalError}, err
//	}
//
//	return &protov1.ResponseWrapper[protov1.SplitPayload]{
//		Status:  protov1.ResultOk,
//		Payload: protov1.SplitPayload(*view),
//	}, nil
//}
//
//func (m *ClientManager) handleSplits(rpc *protov1.RPC) (interface{}, error) {
//
//	var args protov1.SplitsArgs
//	if err := args.PopulateFromRPC(rpc); err != nil {
//		return nil, fmt.Errorf("error parsing splits arguments: %w", err)
//	}
//
//	views, err := m.splitSDK.Splits()
//	if err != nil {
//		return &protov1.ResponseWrapper[protov1.SplitsPayload]{Status: protov1.ResultInternalError}, err
//	}
//
//	var p protov1.SplitsPayload
//	p.Splits = make([]protov1.SplitPayload, 0, len(views))
//	for _, view := range views {
//		p.Splits = append(p.Splits, protov1.SplitPayload(view))
//	}
//
//	response := &protov1.ResponseWrapper[protov1.SplitsPayload]{
//		Status:  protov1.ResultOk,
//		Payload: p,
//	}
//
//	return response, nil
//}

func NewCSM(serializer serializer.Interface, logger logging.LoggerInterface, splitSDK sdk.Interface) *UnBufferedClientStateMachineImpl {
	return &UnBufferedClientStateMachineImpl{
		serializer: serializer,
		logger:     logger,
		splitSDK:   splitSDK,
	}
}

var _ transfer.ClientStateMachine = (*UnBufferedClientStateMachineImpl)(nil)
