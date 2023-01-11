package protocol

import "fmt"

type Version byte

const (
	V1 Version = 0x01
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

type Result byte

const (
	ResultOk Result = 0x00
)

type RPC struct {
	Version Version
	OpCode  OpCode
	Args    []interface{}
}

func ParseArgs[T argsConstraints](r *RPC) (*T, error) {
	var params T

	if !params.isValidFor(r.OpCode) {
		// TODO(mredolatti): BUG!
		return nil, ErrInternal
	}

	return &params, nil
}

const (
	TreatmentArgKeyIdx          int = 0
	TreatmentArgBucketingKeyIdx int = 1
	TreatmentArgFeatureIdx      int = 2
	TreatmentArgAttributesIdx   int = 3
)

type TreatmentArgs struct {
	Key          string
	BucketingKey string
	Feature      string
	Attributes   map[string]interface{}
}

func (t *TreatmentArgs) isValidFor(opode OpCode) bool {
	return opode == OCTreatment || opode == OCTreatmentWithConfig
}

func (t *TreatmentArgs) fromRawArgs(raw []interface{}) error {
	if len(raw) != 4 {
		// TODO(mredolatti) error
	}

	var ok bool

	if t.Key, ok = raw[TreatmentArgKeyIdx].(string); !ok {
		return &InvocationError{
			code: InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing key. expected string, got: %T", raw[TreatmentArgKeyIdx]),
		}
	}

	if t.BucketingKey, ok = raw[TreatmentArgBucketingKeyIdx].(string); !ok {
		return &InvocationError{
			code: InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing bucketing key. expected string, got: %T", raw[TreatmentArgBucketingKeyIdx]),
		}
	}

	if t.Feature, ok = raw[TreatmentArgFeatureIdx].(string); !ok {
		return &InvocationError{
			code: InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing feature. expected string, got: %T", raw[TreatmentArgFeatureIdx]),
		}
	}

	if t.Attributes, ok = raw[TreatmentArgAttributesIdx].(map[string]interface{}); !ok {
		return &InvocationError{
			code: InvocationErrorInvalidArgs,
			message: fmt.Sprintf("error parsing key. expected map[string->object], got: %T", raw[TreatmentArgAttributesIdx]),
		}
	}

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

type Response struct {
	Status    Result
}

type TreatmentResponse struct {
	Response
	Treatment string
}

type TreatmentsResponse struct {
	Response
	Treatments map[string]string
}

type TreatmentWithConfigResponse struct {
	Response
	Treatment string
	Config    string
}

type TreatmentsWithConfigResponse struct {
	Response
	Results map[string]struct {
		Treatment string
		Config    string
	}
}

type argsConstraints interface {
	TreatmentArgs | TreatmentsArgs | TrackArgs
	isValidFor(OpCode) bool
	fromRawArgs(raw []interface{}) error
}
