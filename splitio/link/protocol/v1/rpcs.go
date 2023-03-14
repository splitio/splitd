package v1

import (
	"errors"
	"fmt"

	"github.com/splitio/splitd/splitio/link/protocol"
)

type OpCode byte

const (
	// Registration & login
	OCRegister OpCode = 0x00

	// Treatment-related ops
	OCTreatment            OpCode = 0x11
	OCTreatments           OpCode = 0x12
	OCTreatmentWithConfig  OpCode = 0x13
	OCTreatmentsWithConfig OpCode = 0x14

	// Track-related ops
	OCTrack OpCode = 0x80
)

type RPC struct {
	protocol.RPCBase
	OpCode OpCode
	Args   []interface{}
}

const (
	RegisterArgIDIdx         = 0
	RegisterArgSDKVersionIdx = 1
)

type RegisterArgs struct {
	ID         string
	SDKVersion string
}

func (r *RegisterArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCRegister {
		return ErrIncorrectArguments
	}

	if len(rpc.Args) != 2 {
		// TODO(mredolatti) error
	}

	var ok bool
	if r.ID, ok = rpc.Args[RegisterArgIDIdx].(string); !ok {
		return &InvocationError{
			code:    InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing ID. expected string, got: %T", rpc.Args[RegisterArgIDIdx]),
		}
	}
	if r.SDKVersion, ok = rpc.Args[RegisterArgSDKVersionIdx].(string); !ok {
		return &InvocationError{
			code:    InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing sdk version. expected string, got: %T", rpc.Args[RegisterArgSDKVersionIdx]),
		}
	}

	return nil
}

const (
	TreatmentArgKeyIdx          int = 0
	TreatmentArgBucketingKeyIdx int = 1
	TreatmentArgFeatureIdx      int = 2
	TreatmentArgAttributesIdx   int = 3
)

type TreatmentArgs struct {
	Key          string
	BucketingKey *string
	Feature      string
	Attributes   map[string]interface{}
}

func (t *TreatmentArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCTreatment && rpc.OpCode != OCTreatmentWithConfig {
		return ErrIncorrectArguments
	}
	if len(rpc.Args) != 4 {
		// TODO(mredolatti) error
	}

	var ok bool

	if t.Key, ok = rpc.Args[TreatmentArgKeyIdx].(string); !ok {
		return &InvocationError{
			code:    InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing key. expected string, got: %T", rpc.Args[TreatmentArgKeyIdx]),
		}
	}

	var err error
	if t.BucketingKey, err = getOptionalRef[string](rpc.Args[TreatmentArgBucketingKeyIdx]); err != nil {
		return &InvocationError{
			code:    InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing bucketing key. expected string, got: %T", rpc.Args[TreatmentArgBucketingKeyIdx]),
		}
	}

	if t.Feature, ok = rpc.Args[TreatmentArgFeatureIdx].(string); !ok {
		return &InvocationError{
			code:    InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing feature. expected string, got: %T", rpc.Args[TreatmentArgFeatureIdx]),
		}
	}

	rawAttrs, err := getOptional[map[string]interface{}](rpc.Args[TreatmentArgAttributesIdx])
	if err != nil {
		return &InvocationError{
			code:    InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing attributes. expected map[string->any], got: %T", rpc.Args[TreatmentArgAttributesIdx]),
		}
	}
	t.Attributes = sanitizeAttributes(rawAttrs)

	return nil
}

type TreatmentsArgs struct {
	Key          string
	BucketingKey string
	Feature      string
	Attributes   map[string]interface{}
}

func (t *TreatmentsArgs) isValidFor(opode OpCode) bool {
	return opode == OCTreatments || opode == OCTreatmentsWithConfig
}

func (t *TreatmentsArgs) fromRawArgs(raw []interface{}) error {
	return nil
}

type TrackArgs struct {
	Key         string
	EventType   string
	TrafficType string
	Value       *float64
	Timestamp   int64
}

func (t *TrackArgs) isValidFor(opode OpCode) bool {
	return opode == OCTrack
}

func (t *TrackArgs) fromRawArgs(raw []interface{}) error {
	return nil
}

// -- helpers
var ErrWrongType = errors.New("wrong type")

func getOptionalRef[T any](i interface{}) (*T, error) {

	if i == nil {
		return nil, nil
	}

	ass, ok := i.(T)
	if !ok {
		return nil, ErrWrongType
	}

	return &ass, nil
}

func getOptional[T any /*TODO(mredolatti): restrict!*/](i interface{}) (T, error) {
	if i == nil {
		var t T
		return t, nil
	}

	ass, ok := i.(T)
	if !ok {
		var t T
		return t, ErrWrongType
	}

	return ass, nil
}

func sanitizeAttributes(attrs map[string]interface{}) map[string]interface{} {
	for k, v := range attrs {
		switch parsed := v.(type) {
		case uint8:
			attrs[k] = int(parsed)
		case uint16:
			attrs[k] = int(parsed)
		case uint32:
			attrs[k] = int(parsed)
		case uint64:
			attrs[k] = int(parsed)
		case int8:
			attrs[k] = int(parsed)
		case int16:
			attrs[k] = int(parsed)
		case int32:
			attrs[k] = int(parsed)
		case int64:
			attrs[k] = int(parsed)
		case int:
			attrs[k] = int(parsed)
		case []interface{}:
			asStrSlice := make([]string, len(parsed))
			for idx, item := range parsed {
				if asString, ok := item.(string); ok {
					asStrSlice[idx] = asString
				}
			}
		}
	}
	return attrs
}
