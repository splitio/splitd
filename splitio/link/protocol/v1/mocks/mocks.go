package mocks

import (
	"fmt"
	"os"
	"strconv"

	"github.com/splitio/splitd/splitio"
	"github.com/splitio/splitd/splitio/link/protocol"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/sdk"
)

func NewRegisterRPC(listener bool) *v1.RPC {
	var flags v1.RegisterFlags
	if listener {
		flags = 1 << v1.RegisterFlagReturnImpressionData
	}
	return &v1.RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  v1.OCRegister,
		Args:    []interface{}{strconv.Itoa(os.Getpid()), fmt.Sprintf("splitd-%s", splitio.Version), flags},
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

func NewTreatmentsResp(ok bool, data []sdk.Result) *v1.ResponseWrapper[v1.TreatmentsPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}

	payload := make([]v1.TreatmentPayload, 0, len(data))
	for _, r := range data {
		p := v1.TreatmentPayload{Treatment: r.Treatment}
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
