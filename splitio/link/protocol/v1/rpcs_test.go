package v1

import (
	"testing"
	"time"

	"github.com/splitio/splitd/splitio/common/lang"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/stretchr/testify/assert"
)

func TestOpStringify(t *testing.T) {

	assert.Equal(t, "register", OCRegister.String())
	assert.Equal(t, "treatment", OCTreatment.String())
	assert.Equal(t, "treatment-with-config", OCTreatmentWithConfig.String())
	assert.Equal(t, "treatments", OCTreatments.String())
	assert.Equal(t, "treatments-with-config", OCTreatmentsWithConfig.String())
	assert.Equal(t, "treatments-by-flag-set", OCTreatmentsByFlagSet.String())
	assert.Equal(t, "treatments-with-config-by-flag-set", OCTreatmentsWithConfigByFlagSet.String())
	assert.Equal(t, "treatments-by-flag-sets", OCTreatmentsByFlagSets.String())
	assert.Equal(t, "treatments-with-config-by-flag-sets", OCTreatmentsWithConfigByFlagSets.String())
	assert.Equal(t, "track", OCTrack.String())
	assert.Equal(t, "split-names", OCSplitNames.String())
	assert.Equal(t, "split", OCSplit.String())
	assert.Equal(t, "splits", OCSplits.String())
	assert.Equal(t, "unknown", OpCode(255).String())
}

func TestRegisterRPCParsing(t *testing.T) {
	var r RegisterArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatment, Args: nil}))
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: []interface{}{}}))
	assert.Equal(t, RPCParseError{Code: PECInvalidArgType, Data: int64(RegisterArgIDIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: []interface{}{12, "go-1.2.3", uint64(0)}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(RegisterArgSDKVersionIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: []interface{}{"some", false, uint64(0)}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(RegisterArgFlagsIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: []interface{}{"some", "some_sdk-1.2.3", false}}),
	)

	err := r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCRegister,
		Args:    []interface{}{"some", "some_sdk-1.2.3", uint64(0)},
	})

	assert.Nil(t, err)
	assert.Equal(t, "some", r.ID)
	assert.Equal(t, "some_sdk-1.2.3", r.SDKVersion)
	assert.Equal(t, RegisterFlags(uint64(0)), r.Flags)
}

func TestTreatmentRPCParsing(t *testing.T) {
	var r TreatmentArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatment, Args: []interface{}{}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatment, Args: []interface{}{nil, nil, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgBucketingKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatment, Args: []interface{}{"key", 123, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgFeatureIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatment, Args: []interface{}{"key", "bk", nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentArgAttributesIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatment, Args: []interface{}{"key", "bk", "feat1", 123}}))

	err := r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatment,
		Args:    []interface{}{"key", "bk", "feat1", map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, "feat1", r.Feature)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil bucketing key
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatment,
		Args:    []interface{}{"key", nil, "feat1", map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Nil(t, r.BucketingKey)
	assert.Equal(t, "feat1", r.Feature)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil attributes
	r = TreatmentArgs{}
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatment,
		Args:    []interface{}{"key", "bk", "feat1", nil}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, "feat1", r.Feature)
	assert.Nil(t, r.Attributes)
}

func TestTreatmentsRPCParsing(t *testing.T) {
	var r TreatmentsArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatments, Args: []interface{}{}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatments, Args: []interface{}{nil, nil, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgBucketingKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatments, Args: []interface{}{"key", 123, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgFeaturesIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatments, Args: []interface{}{"key", "bk", 123, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsArgAttributesIdx)},
		r.PopulateFromRPC(&RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  OCTreatments,
			Args:    []interface{}{"key", "bk", []interface{}{"feat1", "feat2"}, 123}}),
	)

	err := r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatments,
		Args:    []interface{}{"key", "bk", []interface{}{"feat1", "feat2"}, map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, []string{"feat1", "feat2"}, r.Features)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil bucketing key
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatments,
		Args:    []interface{}{"key", nil, []interface{}{"feat1", "feat2"}, map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Nil(t, r.BucketingKey)
	assert.Equal(t, []string{"feat1", "feat2"}, r.Features)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatments,
		Args:    []interface{}{"key", "bk", []interface{}{"feat1", "feat2"}, nil}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, []string{"feat1", "feat2"}, r.Features)
	assert.Nil(t, r.Attributes)
}

func TestTreatmentsByFlagSetRPCParsing(t *testing.T) {
	var r TreatmentsByFlagSetArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSet, Args: []interface{}{}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSet, Args: []interface{}{nil, nil, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgBucketingKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSet, Args: []interface{}{"key", 123, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgFlagSetIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSet, Args: []interface{}{"key", "bk", 123, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetArgAttributesIdx)},
		r.PopulateFromRPC(&RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  OCTreatmentsByFlagSet,
			Args:    []interface{}{"key", "bk", "set", 123}}),
	)

	err := r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatmentsByFlagSet,
		Args:    []interface{}{"key", "bk", "set", map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, "set", r.FlagSet)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil bucketing key
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatmentsByFlagSet,
		Args:    []interface{}{"key", nil, "set", map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Nil(t, r.BucketingKey)
	assert.Equal(t, "set", r.FlagSet)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil attributes
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatmentsByFlagSet,
		Args:    []interface{}{"key", "bk", "set", nil}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, "set", r.FlagSet)
	assert.Nil(t, r.Attributes)
}

func TestTreatmentsByFlagSetsRPCParsing(t *testing.T) {
	var r TreatmentsByFlagSetsArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSets, Args: []interface{}{}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSets, Args: []interface{}{nil, nil, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgBucketingKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSets, Args: []interface{}{"key", 123, nil, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgFlagSetsIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTreatmentsByFlagSets, Args: []interface{}{"key", "bk", 123, nil}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TreatmentsByFlagSetsArgAttributesIdx)},
		r.PopulateFromRPC(&RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  OCTreatmentsByFlagSets,
			Args:    []interface{}{"key", "bk", []interface{}{"set_1", "set_2"}, 123}}),
	)

	err := r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatmentsByFlagSets,
		Args:    []interface{}{"key", "bk", []interface{}{"set_1", "set_2"}, map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, []string{"set_1", "set_2"}, r.FlagSets)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil bucketing key
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatmentsByFlagSets,
		Args:    []interface{}{"key", nil, []interface{}{"set_1", "set_2"}, map[string]interface{}{"a": 1}}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Nil(t, r.BucketingKey)
	assert.Equal(t, []string{"set_1", "set_2"}, r.FlagSets)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Attributes)

	// nil attributes
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTreatmentsByFlagSets,
		Args:    []interface{}{"key", "bk", []interface{}{"set_1", "set_2"}, nil}})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, lang.Ref("bk"), r.BucketingKey)
	assert.Equal(t, []string{"set_1", "set_2"}, r.FlagSets)
	assert.Nil(t, r.Attributes)
}

func TestTrackRPCParsing(t *testing.T) {
	var r TrackArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTrack, Args: []interface{}{}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgKeyIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTrack, Args: []interface{}{nil, nil, nil, 123, 123}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgTrafficTypeIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTrack, Args: []interface{}{"key", nil, nil, "asd", 123}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgEventTypeIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTrack, Args: []interface{}{"key", "tt", nil, "asd", 123}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgValueIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTrack, Args: []interface{}{"key", "tt", "et", "asd", 123}}))

	assert.Equal(t,
		RPCParseError{Code: PECInvalidArgType, Data: int64(TrackArgPropertiesIdx)},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCTrack, Args: []interface{}{"key", "tt", "et", 2.8, 123}}))

	err := r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTrack,
		Args:    []interface{}{"key", "tt", "et", 2.8, map[string]interface{}{"a": int64(1)}},
	})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, "tt", r.TrafficType)
	assert.Equal(t, "et", r.EventType)
	assert.Equal(t, lang.Ref(float64(2.8)), r.Value)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Properties)

	// nil properties
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTrack,
		Args:    []interface{}{"key", "tt", "et", 2.8, nil},
	})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, "tt", r.TrafficType)
	assert.Equal(t, "et", r.EventType)
	assert.Equal(t, lang.Ref(float64(2.8)), r.Value)
	assert.Nil(t, r.Properties)

	// nil value
	r = TrackArgs{}
	err = r.PopulateFromRPC(&RPC{
		RPCBase: protocol.RPCBase{Version: protocol.V1},
		OpCode:  OCTrack,
		Args:    []interface{}{"key", "tt", "et", nil, map[string]interface{}{"a": int64(1)}},
	})
	assert.Nil(t, err)
	assert.Equal(t, "key", r.Key)
	assert.Equal(t, "tt", r.TrafficType)
	assert.Equal(t, "et", r.EventType)
	assert.Nil(t, r.Value)
	assert.Equal(t, map[string]interface{}{"a": int64(1)}, r.Properties)

}

func TestSplitNamesRPCProcessing(t *testing.T) {
	var r SplitNamesArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplitNames, Args: []interface{}{"asd"}}),
	)

	err := r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplitNames, Args: []interface{}{}})
	assert.Nil(t, err)

	err = r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplitNames, Args: nil})
	assert.Nil(t, err)
}

func TestSplitsRPCProcessing(t *testing.T) {
	var r SplitsArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: nil}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplits, Args: []interface{}{"asd"}}),
	)

	err := r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplits, Args: []interface{}{}})
	assert.Nil(t, err)

	err = r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplits, Args: nil})
	assert.Nil(t, err)
}

func TestSplitRPCProcessing(t *testing.T) {
	var r SplitArgs
	assert.Equal(t,
		RPCParseError{Code: PECOpCodeMismatch},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCRegister, Args: []interface{}{"s1"}}),
	)
	assert.Equal(t,
		RPCParseError{Code: PECWrongArgCount},
		r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplit, Args: []interface{}{"s1", "s2"}}),
	)

	err := r.PopulateFromRPC(&RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: OCSplit, Args: []interface{}{"s1"}})
	assert.Nil(t, err)
	assert.Equal(t, "s1", r.Name)
}

func TestSanitizeAttributes(t *testing.T) {
	now := time.Now()
	attrs := map[string]interface{}{
		"s":      "some string",
		"i":      int(1),
		"i8":     int8(2),
		"i16":    int16(3),
		"i32":    int32(4),
		"i64":    int64(5),
		"u8":     uint8(6),
		"u16":    uint16(7),
		"u32":    uint32(8),
		"u64":    uint64(9),
		"allStr": []interface{}{"a", "b", "c"},
		"mixed":  []interface{}{"a", 1, true},
		"time":   now,
	}

	attrs = sanitizeAttributes(attrs)
	assert.Equal(t, "some string", attrs["s"])
	assert.Equal(t, int64(1), attrs["i"])
	assert.Equal(t, int64(2), attrs["i8"])
	assert.Equal(t, int64(3), attrs["i16"])
	assert.Equal(t, int64(4), attrs["i32"])
	assert.Equal(t, int64(5), attrs["i64"])
	assert.Equal(t, int64(6), attrs["u8"])
	assert.Equal(t, int64(7), attrs["u16"])
	assert.Equal(t, int64(8), attrs["u32"])
	assert.Equal(t, int64(9), attrs["u64"])
	assert.Equal(t, []string{"a", "b", "c"}, attrs["allStr"])
	assert.Equal(t, []string{"a"}, attrs["mixed"])
	assert.Equal(t, now.Unix(), attrs["time"])
}

func TestRPCEncoding(t *testing.T) {
	ra := RegisterArgs{
		ID:         "someID",
		SDKVersion: "some-1.2.3",
		Flags:      0,
	}
	encodedRA := ra.Encode()
	assert.Equal(t, ra.ID, encodedRA[RegisterArgIDIdx].(string))
	assert.Equal(t, ra.SDKVersion, encodedRA[RegisterArgSDKVersionIdx].(string))
	assert.Equal(t, ra.Flags, encodedRA[RegisterArgFlagsIdx].(RegisterFlags))

	ta := TreatmentArgs{
		Key:          "someKey",
		BucketingKey: lang.Ref("someBucketing"),
		Feature:      "someFeature",
		Attributes:   map[string]interface{}{"some": "attribute"},
	}
	encodedTA := ta.Encode()
	assert.Equal(t, ta.Key, encodedTA[TreatmentArgKeyIdx].(string))
	assert.Equal(t, *ta.BucketingKey, encodedTA[TreatmentArgBucketingKeyIdx].(string))
	assert.Equal(t, ta.Feature, encodedTA[TreatmentArgFeatureIdx].(string))
	assert.Equal(t, ta.Attributes, encodedTA[TreatmentArgAttributesIdx].(map[string]interface{}))

	tsa := TreatmentsArgs{
		Key:          "someKey",
		BucketingKey: lang.Ref("someBucketing"),
		Features:     []string{"someFeature", "someFeature2"},
		Attributes:   map[string]interface{}{"some": "attribute"},
	}
	encodedTsA := tsa.Encode()
	assert.Equal(t, tsa.Key, encodedTsA[TreatmentsArgKeyIdx].(string))
	assert.Equal(t, *tsa.BucketingKey, encodedTsA[TreatmentsArgBucketingKeyIdx].(string))
	assert.Equal(t, tsa.Features, encodedTsA[TreatmentsArgFeaturesIdx].([]string))
	assert.Equal(t, tsa.Attributes, encodedTsA[TreatmentsArgAttributesIdx].(map[string]interface{}))

	tsfa := TreatmentsByFlagSetArgs{
		Key:          "someKey",
		BucketingKey: lang.Ref("someBucketing"),
		FlagSet:      "set",
		Attributes:   map[string]interface{}{"some": "attribute"},
	}
	encodedTsfA := tsfa.Encode()
	assert.Equal(t, tsfa.Key, encodedTsfA[TreatmentsByFlagSetArgKeyIdx].(string))
	assert.Equal(t, *tsfa.BucketingKey, encodedTsfA[TreatmentsByFlagSetArgBucketingKeyIdx].(string))
	assert.Equal(t, tsfa.FlagSet, encodedTsfA[TreatmentsByFlagSetArgFlagSetIdx].(string))
	assert.Equal(t, tsfa.Attributes, encodedTsfA[TreatmentsByFlagSetArgAttributesIdx].(map[string]interface{}))

	tsfsa := TreatmentsByFlagSetsArgs{
		Key:          "someKey",
		BucketingKey: lang.Ref("someBucketing"),
		FlagSets:     []string{"set_1", "set_2"},
		Attributes:   map[string]interface{}{"some": "attribute"},
	}
	encodedTsfsA := tsfsa.Encode()
	assert.Equal(t, tsfsa.Key, encodedTsfsA[TreatmentsByFlagSetsArgKeyIdx].(string))
	assert.Equal(t, *tsfsa.BucketingKey, encodedTsfsA[TreatmentsByFlagSetsArgBucketingKeyIdx].(string))
	assert.Equal(t, tsfsa.FlagSets, encodedTsfsA[TreatmentsByFlagSetsArgFlagSetsIdx].([]string))
	assert.Equal(t, tsfsa.Attributes, encodedTsfsA[TreatmentsByFlagSetsArgAttributesIdx].(map[string]interface{}))

	tra := TrackArgs{
		Key:         "someKey",
		TrafficType: "someTrafficType",
		EventType:   "someEventType",
		Value:       lang.Ref(123.),
		Properties:  map[string]interface{}{"a": 1},
	}
	encodedTrA := tra.Encode()
	assert.Equal(t, tra.Key, encodedTrA[TrackArgKeyIdx].(string))
	assert.Equal(t, tra.TrafficType, encodedTrA[TrackArgTrafficTypeIdx].(string))
	assert.Equal(t, tra.EventType, encodedTrA[TrackArgEventTypeIdx].(string))
	assert.Equal(t, *tra.Value, encodedTrA[TrackArgValueIdx].(float64))
	assert.Equal(t, tra.Properties, encodedTrA[TrackArgPropertiesIdx].(map[string]interface{}))
}
