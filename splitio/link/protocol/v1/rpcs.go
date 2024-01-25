package v1

import (
	"errors"
	"time"

	"github.com/splitio/splitd/splitio/link/protocol"
)

type OpCode byte

const (
	// Registration & login
	OCRegister OpCode = 0x00

	// Treatment-related ops
	OCTreatment                      OpCode = 0x11
	OCTreatments                     OpCode = 0x12
	OCTreatmentWithConfig            OpCode = 0x13
	OCTreatmentsWithConfig           OpCode = 0x14
	OCTreatmentsByFlagSet            OpCode = 0x15
	OCTreatmentsWithConfigByFlagSet  OpCode = 0x16
	OCTreatmentsByFlagSets           OpCode = 0x17
	OCTreatmentsWithConfigByFlagSets OpCode = 0x18

	// Track-related ops
	OCTrack OpCode = 0x80

	OCSplitNames OpCode = 0xA0
	OCSplit      OpCode = 0xA1
	OCSplits     OpCode = 0xA2
)

func (o OpCode) String() string {
	switch o {
	case OCRegister:
		return "register"
	case OCTreatment:
		return "treatment"
	case OCTreatments:
		return "treatments"
	case OCTreatmentWithConfig:
		return "treatment-with-config"
	case OCTreatmentsWithConfig:
		return "treatments-with-config"
	case OCTreatmentsByFlagSet:
		return "treatments-by-flag-set"
	case OCTreatmentsWithConfigByFlagSet:
		return "treatments-with-config-by-flag-set"
	case OCTreatmentsByFlagSets:
		return "treatments-by-flag-sets"
	case OCTreatmentsWithConfigByFlagSets:
		return "treatments-with-config-by-flag-sets"
	case OCTrack:
		return "track"
	case OCSplitNames:
		return "split-names"
	case OCSplit:
		return "split"
	case OCSplits:
		return "splits"
	default:
		return "unknown"
	}
}

type RPC struct {
	protocol.RPCBase
	OpCode OpCode        `msgpack:"o"`
	Args   []interface{} `msgpack:"a"`
}

type Arguments interface {
	PopulateFromRPC(rpc *RPC) error
	Encode() []interface{}
}

type RegisterArgs struct {
	ID         string        `msgpack:"i"`
	SDKVersion string        `msgpack:"s"`
	Flags      RegisterFlags `msgpack:"f"`
}

const (
	RegisterArgIDIdx         = 0
	RegisterArgSDKVersionIdx = 1
	RegisterArgFlagsIdx      = 2
)

type RegisterFlags uint64

const (
	RegisterFlagReturnImpressionData RegisterFlags = (1 << 0)
)

func (r RegisterArgs) Encode() []interface{} {
	return []interface{}{r.ID, r.SDKVersion, r.Flags}
}

func (r *RegisterArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCRegister {
		return RPCParseError{Code: PECOpCodeMismatch}
	}

	if len(rpc.Args) != 3 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	if r.ID, ok = rpc.Args[RegisterArgIDIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(RegisterArgIDIdx)}
	}
	if r.SDKVersion, ok = rpc.Args[RegisterArgSDKVersionIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(RegisterArgSDKVersionIdx)}
	}
	if asUInt, ok := tryInt[uint64](rpc.Args[RegisterArgFlagsIdx]); ok {
		r.Flags = RegisterFlags(asUInt)
	} else {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(RegisterArgFlagsIdx)}
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
	Key          string                 `msgpack:"k"`
	BucketingKey *string                `msgpack:"b"`
	Feature      string                 `msgpack:"f"`
	Attributes   map[string]interface{} `msgpack:"a"`
}

func (r TreatmentArgs) Encode() []interface{} {
	var bk string
	if r.BucketingKey != nil {
		bk = *r.BucketingKey
	}
	return []interface{}{r.Key, bk, r.Feature, r.Attributes}
}

func (t *TreatmentArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCTreatment && rpc.OpCode != OCTreatmentWithConfig {
		return RPCParseError{Code: PECOpCodeMismatch}
	}
	if len(rpc.Args) != 4 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	var err error

	if t.Key, ok = rpc.Args[TreatmentArgKeyIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgKeyIdx)}
	}

	if t.BucketingKey, err = getOptionalRef[string](rpc.Args[TreatmentArgBucketingKeyIdx]); err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgBucketingKeyIdx)}

	}

	if t.Feature, ok = rpc.Args[TreatmentArgFeatureIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgFeatureIdx)}

	}

	if rpc.Args[TreatmentArgAttributesIdx] != nil {
		rawAttrs, err := getOptional[map[string]interface{}](rpc.Args[TreatmentArgAttributesIdx])
		if err != nil {
			return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgAttributesIdx)}
		}
		t.Attributes = sanitizeAttributes(rawAttrs)
	}

	return nil
}

const (
	TreatmentsArgKeyIdx          int = 0
	TreatmentsArgBucketingKeyIdx int = 1
	TreatmentsArgFeaturesIdx     int = 2
	TreatmentsArgAttributesIdx   int = 3
)

type TreatmentsArgs struct {
	Key          string                 `msgpack:"k"`
	BucketingKey *string                `msgpack:"b"`
	Features     []string               `msgpack:"f"`
	Attributes   map[string]interface{} `msgpack:"a"`
}

func (r TreatmentsArgs) Encode() []interface{} {
	var bk string
	if r.BucketingKey != nil {
		bk = *r.BucketingKey
	}
	return []interface{}{r.Key, bk, r.Features, r.Attributes}
}

func (t *TreatmentsArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCTreatments && rpc.OpCode != OCTreatmentsWithConfig {
		return RPCParseError{Code: PECOpCodeMismatch}
	}
	if len(rpc.Args) != 4 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	var err error

	if t.Key, ok = rpc.Args[TreatmentsArgKeyIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgKeyIdx)}
	}

	if t.BucketingKey, err = getOptionalRef[string](rpc.Args[TreatmentsArgBucketingKeyIdx]); err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgBucketingKeyIdx)}

	}

	rawFeatureList, ok := rpc.Args[TreatmentsArgFeaturesIdx].([]interface{})
	if !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgFeaturesIdx)}

	}
	t.Features, ok = sanitizeToStringSlice(rawFeatureList)
	if !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgFeaturesIdx)}
	}

	rawAttrs, err := getOptional[map[string]interface{}](rpc.Args[TreatmentsArgAttributesIdx])
	if err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgAttributesIdx)}
	}
	t.Attributes = sanitizeAttributes(rawAttrs)

	return nil
}

const (
	TreatmentsByFlagSetArgKeyIdx          int = 0
	TreatmentsByFlagSetArgBucketingKeyIdx int = 1
	TreatmentsByFlagSetArgFlagSetIdx      int = 2
	TreatmentsByFlagSetArgAttributesIdx   int = 3
)

type TreatmentsByFlagSetArgs struct {
	Key          string                 `msgpack:"k"`
	BucketingKey *string                `msgpack:"b"`
	FlagSet      string                 `msgpack:"f"`
	Attributes   map[string]interface{} `msgpack:"a"`
}

func (r TreatmentsByFlagSetArgs) Encode() []interface{} {
	var bk string
	if r.BucketingKey != nil {
		bk = *r.BucketingKey
	}
	return []interface{}{r.Key, bk, r.FlagSet, r.Attributes}
}

func (t *TreatmentsByFlagSetArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCTreatmentsByFlagSet && rpc.OpCode != OCTreatmentsWithConfigByFlagSet {
		return RPCParseError{Code: PECOpCodeMismatch}
	}
	if len(rpc.Args) != 4 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	var err error

	if t.Key, ok = rpc.Args[TreatmentsByFlagSetArgKeyIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgKeyIdx)}
	}

	if t.BucketingKey, err = getOptionalRef[string](rpc.Args[TreatmentsByFlagSetArgBucketingKeyIdx]); err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgBucketingKeyIdx)}
	}

	if t.FlagSet, ok = rpc.Args[TreatmentsByFlagSetArgFlagSetIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgFlagSetIdx)}
	}

	rawAttrs, err := getOptional[map[string]interface{}](rpc.Args[TreatmentsByFlagSetArgAttributesIdx])
	if err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgAttributesIdx)}
	}
	t.Attributes = sanitizeAttributes(rawAttrs)

	return nil
}

const (
	TreatmentsByFlagSetsArgKeyIdx          int = 0
	TreatmentsByFlagSetsArgBucketingKeyIdx int = 1
	TreatmentsByFlagSetsArgFlagSetsIdx     int = 2
	TreatmentsByFlagSetsArgAttributesIdx   int = 3
)

type TreatmentsByFlagSetsArgs struct {
	Key          string                 `msgpack:"k"`
	BucketingKey *string                `msgpack:"b"`
	FlagSets     []string               `msgpack:"f"`
	Attributes   map[string]interface{} `msgpack:"a"`
}

func (r TreatmentsByFlagSetsArgs) Encode() []interface{} {
	var bk string
	if r.BucketingKey != nil {
		bk = *r.BucketingKey
	}
	return []interface{}{r.Key, bk, r.FlagSets, r.Attributes}
}

func (t *TreatmentsByFlagSetsArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCTreatmentsByFlagSets && rpc.OpCode != OCTreatmentsWithConfigByFlagSets {
		return RPCParseError{Code: PECOpCodeMismatch}
	}
	if len(rpc.Args) != 4 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	var err error

	if t.Key, ok = rpc.Args[TreatmentsByFlagSetsArgKeyIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgKeyIdx)}
	}

	if t.BucketingKey, err = getOptionalRef[string](rpc.Args[TreatmentsByFlagSetsArgBucketingKeyIdx]); err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgBucketingKeyIdx)}
	}

	rawFlagSetsList, ok := rpc.Args[TreatmentsByFlagSetsArgFlagSetsIdx].([]interface{})
	if !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgFlagSetsIdx)}
	}

	t.FlagSets, ok = sanitizeToStringSlice(rawFlagSetsList)
	if !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgFlagSetsIdx)}
	}

	rawAttrs, err := getOptional[map[string]interface{}](rpc.Args[TreatmentsByFlagSetsArgAttributesIdx])
	if err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgAttributesIdx)}
	}
	t.Attributes = sanitizeAttributes(rawAttrs)

	return nil
}

const (
	TrackArgKeyIdx         int = 0
	TrackArgTrafficTypeIdx int = 1
	TrackArgEventTypeIdx   int = 2
	TrackArgValueIdx       int = 3
	TrackArgPropertiesIdx  int = 4
)

type TrackArgs struct {
	Key         string                 `msgpack:"k"`
	TrafficType string                 `msgpack:"t"`
	EventType   string                 `msgpack:"e"`
	Value       *float64               `msgpack:"v"`
	Properties  map[string]interface{} `msgpack:"p"`
}

func (r TrackArgs) Encode() []interface{} {
	asInterface := make([]interface{}, 0, 5)
	asInterface = append(asInterface, r.Key, r.TrafficType, r.EventType)
	if r.Value == nil {
		asInterface = append(asInterface, nil)
	}
	asInterface = append(asInterface, *r.Value)
	asInterface = append(asInterface, r.Properties)
	return asInterface
}

func (t *TrackArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCTrack {
		return RPCParseError{Code: PECOpCodeMismatch}
	}
	if len(rpc.Args) != 5 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	var err error

	if t.Key, ok = rpc.Args[TrackArgKeyIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgKeyIdx)}
	}

	if t.TrafficType, ok = rpc.Args[TrackArgTrafficTypeIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgTrafficTypeIdx)}
	}

	if t.EventType, ok = rpc.Args[TrackArgEventTypeIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgEventTypeIdx)}
	}

	if rpc.Args[TrackArgValueIdx] != nil {
		if val, ok := tryNumberAsFloat(rpc.Args[TrackArgValueIdx]); ok {
			t.Value = &val
		} else {
			return RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgValueIdx)}
		}
	}

	if t.Properties, err = getOptional[map[string]interface{}](rpc.Args[TrackArgPropertiesIdx]); err != nil {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgPropertiesIdx)}
	}

	return nil
}

type SplitNamesArgs struct{}

func (s SplitNamesArgs) Encode() []interface{} {
	return nil
}

func (t *SplitNamesArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCSplitNames {
		return RPCParseError{Code: PECOpCodeMismatch}
	}

	if len(rpc.Args) != 0 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	return nil
}

type SplitsArgs struct{}

func (s SplitsArgs) Encode() []interface{} {
	return nil
}

func (t *SplitsArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCSplits {
		return RPCParseError{Code: PECOpCodeMismatch}
	}

	if len(rpc.Args) != 0 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	return nil
}

const (
	SplitArgNameIdx int = 0
)

type SplitArgs struct {
	Name string
}

func (s SplitArgs) Encode() []interface{} {
	return []interface{}{s.Name}
}

func (t *SplitArgs) PopulateFromRPC(rpc *RPC) error {
	if rpc.OpCode != OCSplit {
		return RPCParseError{Code: PECOpCodeMismatch}
	}

	if len(rpc.Args) != 1 {
		return RPCParseError{Code: PECWrongArgCount}
	}

	var ok bool
	if t.Name, ok = rpc.Args[SplitArgNameIdx].(string); !ok {
		return RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgKeyIdx)}
	}

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

		if asInt, ok := tryInt[int64](v); ok {
			attrs[k] = asInt
		}

		switch parsed := v.(type) {
		case time.Time:
			attrs[k] = parsed.Unix()
		case []interface{}:
			asStrSlice := make([]string, len(parsed))
			var added int
			for _, item := range parsed {
				if asString, ok := item.(string); ok {
					asStrSlice[added] = asString
					added++
				}
			}
			attrs[k] = asStrSlice[:added]
		}
	}
	return attrs
}

func sanitizeToStringSlice(raw []interface{}) ([]string, bool) {
	asStringSlice := make([]string, 0, len(raw))
	for _, f := range raw {
		asStr, ok := f.(string)
		if !ok {
			return nil, false
		}
		asStringSlice = append(asStringSlice, asStr)
	}
	return asStringSlice, true
}

func tryInt[T int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](x interface{}) (T, bool) {
	switch parsed := x.(type) {
	case uint8:
		return T(parsed), true
	case uint16:
		return T(parsed), true
	case uint32:
		return T(parsed), true
	case uint64:
		return T(parsed), true
	case int8:
		return T(parsed), true
	case int16:
		return T(parsed), true
	case int32:
		return T(parsed), true
	case int64:
		return T(parsed), true
	case int:
		return T(parsed), true
	case uint:
		return T(parsed), true
	}
	return T(0), false
}

func tryNumberAsFloat(x interface{}) (float64, bool) {
	if asInt, ok := tryInt[int64](x); ok {
		return float64(asInt), true
	}

	switch parsed := x.(type) {
	case float32:
		return float64(parsed), true
	case float64:
		return parsed, true
	}

	return 0, false
}

var _ Arguments = (*RegisterArgs)(nil)
var _ Arguments = (*TreatmentArgs)(nil)
var _ Arguments = (*TreatmentsArgs)(nil)
var _ Arguments = (*TrackArgs)(nil)
