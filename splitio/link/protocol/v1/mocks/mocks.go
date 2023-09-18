package mocks

import (
	"fmt"

	"github.com/splitio/splitd/splitio"
	"github.com/splitio/splitd/splitio/common/lang"
	"github.com/splitio/splitd/splitio/link/protocol"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/sdk"
)

func NewRegisterRPC(id string, listener bool) *v1.RPC {
	var flags v1.RegisterFlags
	if listener {
		flags = 1 << v1.RegisterFlagReturnImpressionData
	}
	return &v1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  v1.OCRegister,
		Args:    []interface{}{id, fmt.Sprintf("splitd-%s", splitio.Version), flags},
	}
}

func NewTreatmentRPC(key string, bucketing string, feature string, attrs map[string]interface{}) *v1.RPC {
	return &v1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  v1.OCTreatment,
		Args:    []interface{}{key, bucketing, feature, attrs},
	}
}

func NewTreatmentsRPC(key string, bucketing string, features []string, attrs map[string]interface{}) *v1.RPC {
	return &v1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  v1.OCTreatments,
		Args:    []interface{}{key, bucketing, features, attrs},
	}
}

func NewTrackRPC(key string, trafficType string, eventType string, eventVal *float64, props map[string]interface{}) *v1.RPC {
	return &v1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  v1.OCTrack,
		Args:    []interface{}{key, trafficType, eventType, nilOrVal(eventVal), props},
	}
}

func NewSplitNamesRPC() *v1.RPC {
	return &v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCSplitNames}
}

func NewSplitsRPC() *v1.RPC {
	return &v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCSplits}
}

func NewSplitRPC(name string) *v1.RPC {
	return &v1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  v1.OCSplit,
		Args:    []interface{}{name},
	}
}

func NewRegisterResp(ok bool) *v1.ResponseWrapper[v1.RegisterPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.RegisterPayload]{
		Status:  res,
		Payload: v1.RegisterPayload{},
	}
}

func NewTreatmentResp(ok bool, treatment string, ilData *v1.ListenerExtraData) *v1.ResponseWrapper[v1.TreatmentPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.TreatmentPayload]{
		Status: res,
		Payload: v1.TreatmentPayload{
			Treatment:    treatment,
			ListenerData: ilData,
		},
	}
}

func NewTreatmentWithConfigResp(ok bool, treatment string, ilData *v1.ListenerExtraData, cfg string) *v1.ResponseWrapper[v1.TreatmentPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.TreatmentPayload]{
		Status: res,
		Payload: v1.TreatmentPayload{
			Treatment:    treatment,
			ListenerData: ilData,
			Config:       lang.Ref(cfg),
		},
	}
}

func NewTreatmentsResp(ok bool, data []sdk.EvaluationResult) *v1.ResponseWrapper[v1.TreatmentsPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}

	payload := make([]v1.TreatmentPayload, 0, len(data))
	for _, r := range data {
		p := v1.TreatmentPayload{
			Treatment: r.Treatment,
			Config:    r.Config,
		}
		if r.Impression != nil {
			p.ListenerData = &v1.ListenerExtraData{
				Label:        r.Impression.Label,
				Timestamp:    r.Impression.Time,
				ChangeNumber: r.Impression.ChangeNumber,
			}
		}
		payload = append(payload, p)
	}

	return &v1.ResponseWrapper[v1.TreatmentsPayload]{
		Status:  res,
		Payload: v1.TreatmentsPayload{Results: payload},
	}
}

func NewTrackResp(ok bool) *v1.ResponseWrapper[v1.TrackPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.TrackPayload]{
		Status:  res,
		Payload: v1.TrackPayload{Success: ok},
	}
}

func NewSplitNamesResp(ok bool, names []string) *v1.ResponseWrapper[v1.SplitNamesPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.SplitNamesPayload]{
		Status:  res,
		Payload: v1.SplitNamesPayload{Names: names},
	}
}

func NewSplitResp(ok bool, split v1.SplitPayload) *v1.ResponseWrapper[v1.SplitPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.SplitPayload]{Status: res, Payload: split}
}

func NewSplitsResp(ok bool, splits []v1.SplitPayload) *v1.ResponseWrapper[v1.SplitsPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.SplitsPayload]{
		Status:  res,
		Payload: v1.SplitsPayload{Splits: splits},
	}
}

func nilOrVal(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
