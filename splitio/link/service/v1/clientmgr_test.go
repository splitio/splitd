package v1

import (
	"errors"
	"io"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/common/lang"
	"github.com/splitio/splitd/splitio/link/protocol"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	proto1Mocks "github.com/splitio/splitd/splitio/link/protocol/v1/mocks"
	serializerMocks "github.com/splitio/splitd/splitio/link/serializer/mocks"
	transferMocks "github.com/splitio/splitd/splitio/link/transfer/mocks"
	"github.com/splitio/splitd/splitio/sdk"
	sdkMocks "github.com/splitio/splitd/splitio/sdk/mocks"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterAndTreatmentHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatment,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewTreatmentResp(true, "on", nil)).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.
		On("Treatment", &types.ClientConfig{Metadata: types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}}, "key", (*string)(nil), "someFeature", map[string]interface{}(nil)).
		Return(&sdk.EvaluationResult{Treatment: "on"}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestRegisterAndTreatmentsHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentsMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentsMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatments,
			Args:    []interface{}{"key", nil, []interface{}{"feat1", "feat2", "feat3"}, map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsResp(true, []sdk.EvaluationResult{
		{Treatment: "on"}, {Treatment: "off"}, {Treatment: "control"},
	})).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.
		On(
			"Treatments",
			&types.ClientConfig{Metadata: types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}},
			"key",
			(*string)(nil),
			[]string{"feat1", "feat2", "feat3"},
			map[string]interface{}(nil),
		).Return(map[string]sdk.EvaluationResult{
		"feat1": {Treatment: "on"},
		"feat2": {Treatment: "off"},
		"feat3": {Treatment: "control"},
	}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestRegisterAndTreatmentWithConfigHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentWithConfigMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentWithConfigMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatmentWithConfig,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewTreatmentWithConfigResp(true, "on", nil, `{"a": "some"}`)).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.
		On("Treatment",
			&types.ClientConfig{Metadata: types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}},
			"key", (*string)(nil), "someFeature", map[string]interface{}(nil)).
		Return(&sdk.EvaluationResult{Treatment: "on", Config: lang.Ref(`{"a": "some"}`)}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestRegisterAndTreatmentsWithConfigHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentsWithConfigMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	var strCfg = "what"

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentsWithConfigMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatmentsWithConfig,
			Args:    []interface{}{"key", nil, []interface{}{"feat1", "feat2", "feat3"}, map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsResp(true, []sdk.EvaluationResult{
		{Treatment: "on"}, {Treatment: "off"}, {Treatment: "control", Config: &strCfg},
	})).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.
		On(
			"Treatments",
			&types.ClientConfig{Metadata: types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}},
			"key",
			(*string)(nil),
			[]string{"feat1", "feat2", "feat3"},
			map[string]interface{}(nil),
		).Return(map[string]sdk.EvaluationResult{
		"feat1": {Treatment: "on"},
		"feat2": {Treatment: "off"},
		"feat3": {Treatment: "control", Config: &strCfg},
	}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestRegisterWithImpsAndTreatmentHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(v1.RegisterFlagReturnImpressionData)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatment,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewTreatmentResp(true, "on", &v1.ListenerExtraData{Label: "l1", Timestamp: 1234556})).
		Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.
		On("Treatment",
			&types.ClientConfig{Metadata: types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}, ReturnImpressionData: true},
			"key", (*string)(nil), "someFeature", map[string]interface{}(nil)).
		Return(&sdk.EvaluationResult{Treatment: "on", Impression: &dtos.Impression{Label: "l1", Time: 1234556}}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestTrack(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("trackMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("trackMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = *proto1Mocks.NewTrackRPC("key1", "user", "checkin", lang.Ref(2.75), map[string]interface{}{"a": 1})
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewTrackResp(true)).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.
		On("Track",
			&types.ClientConfig{Metadata: types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}},
			"key1", "user", "checkin", lang.Ref(float64(2.75)), map[string]interface{}{"a": 1}).
		Return((error)(nil)).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestSplitNames(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("splitNames"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("splitNames"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = *proto1Mocks.NewSplitNamesRPC()
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewSplitNamesResp(true, []string{"split1", "split2"})).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.On("SplitNames").Return([]string{"split1", "split2"}, (error)(nil)).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestSplits(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("splits"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("splits"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = *proto1Mocks.NewSplitsRPC()
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewSplitsResp(true, []v1.SplitPayload{
		{Name: "s1", TrafficType: "tt1", Killed: true, Treatments: []string{"on", "off"}, ChangeNumber: 1, Configs: map[string]string{"a": "x"}},
		{Name: "s2", TrafficType: "tt1", Killed: false, Treatments: []string{"on", "off"}, ChangeNumber: 1, Configs: map[string]string{"a": "y"}},
	})).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.On("Splits").Return([]sdk.SplitView{
		{Name: "s1", TrafficType: "tt1", Killed: true, Treatments: []string{"on", "off"}, ChangeNumber: 1, Configs: map[string]string{"a": "x"}},
		{Name: "s2", TrafficType: "tt1", Killed: false, Treatments: []string{"on", "off"}, ChangeNumber: 1, Configs: map[string]string{"a": "y"}},
	}, (error)(nil)).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestSplit(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("split"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("split"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = *proto1Mocks.NewSplitRPC("s1")
	}).Once()
	serializerMock.On("Serialize", proto1Mocks.NewSplitResp(true, v1.SplitPayload{
		Name:         "s1",
		TrafficType:  "tt1",
		Killed:       true,
		Treatments:   []string{"on", "off"},
		ChangeNumber: 1,
		Configs:      map[string]string{"a": "x"},
	})).Return([]byte("successPayload"), nil).Once()

	sdkMock := &sdkMocks.SDKMock{}
	sdkMock.On("Split", "s1").Return(&sdk.SplitView{
		Name:         "s1",
		TrafficType:  "tt1",
		Killed:       true,
		Treatments:   []string{"on", "off"},
		ChangeNumber: 1,
		Configs:      map[string]string{"a": "x"},
	}, (error)(nil)).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
}

func TestTreatmentWithoutRegister(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", []byte("treatmentMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatment,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()

	sdkMock := &sdkMocks.SDKMock{}
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Contains(t, err.Error(), "first call must be 'register'")
}

func TestConnectionFailureWhenReading(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), errors.New("something")).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	sdkMock := &sdkMocks.SDKMock{}
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Contains(t, err.Error(), "error reading from conn")
}

func TestManagePanicRecovers(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Panic("some panic")

	logger := &loggerMock{}
	logger.On("Error", "CRITICAL - connection handler is panicking: ", "some panic").Once()
	logger.On("Error", mock.AnythingOfType("string")).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	sdkMock := &sdkMocks.SDKMock{}

	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	cm.Manage()

	logger.AssertExpectations(t)
}

func TestFetchRPC(t *testing.T) {
	// error reading from conn
	someErr := errors.New("someConnErr")
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), someErr)
	serializerMock := &serializerMocks.SerializerMock{}
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, nil, serializerMock)
	rpc, err := cm.fetchRPC()
	assert.Nil(t, rpc)
	assert.ErrorContains(t, err, "someConnErr")

	// error parsing message
	someErr = errors.New("someSerializationErr")
	rawConnMock = &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte{}, nil)
	serializerMock = &serializerMocks.SerializerMock{}
	serializerMock.On("Parse", mock.Anything, mock.Anything).Return(someErr)
	cm = NewClientManager(rawConnMock, logger, nil, serializerMock)
	rpc, err = cm.fetchRPC()
	assert.Nil(t, rpc)
	assert.ErrorContains(t, err, "someSerializationErr")
}

func TestSendResponse(t *testing.T) {
	// error parsing message
	someErr := errors.New("someSerializationErr")
	rawConnMock := &transferMocks.RawConnMock{}
	serializerMock := &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", mock.Anything).Return([]byte(nil), someErr)
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, nil, serializerMock)
	err := cm.sendResponse(nil)
	assert.ErrorContains(t, err, "someSerializationErr")

	// error reading from conn
	someErr = errors.New("someConnErr")
	rawConnMock = &transferMocks.RawConnMock{}
	rawConnMock.On("SendMessage", mock.Anything).Return(someErr)
	serializerMock = &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", mock.Anything).Return([]byte{}, nil)
	cm = NewClientManager(rawConnMock, logger, nil, serializerMock)
	err = cm.sendResponse(nil)
	assert.ErrorContains(t, err, "someConnErr")
}

func TestHandleRPCErrors(t *testing.T) {
	logger := logging.NewLogger(nil)
	cm := NewClientManager(nil, logger, nil, nil)
	res, err := cm.dispatchRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCTreatment})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "first call must be 'register'")

	// register wrong args
	res, err = cm.dispatchRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCRegister, Args: []interface{}{1, "hola"}})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "error parsing register arguments")

	// set the config to allow other rpcs to be handled
	cm.clientConfig = &types.ClientConfig{ReturnImpressionData: true}

	// treatment wrong args
	res, err = cm.dispatchRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCTreatment, Args: []interface{}{1, "hola"}})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "error parsing treatment arguments")

	// register wrong args
	res, err = cm.dispatchRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCTreatments, Args: []interface{}{1, "hola"}})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "error parsing treatments arguments")
}

type loggerMock struct{ mock.Mock }

func (m *loggerMock) Debug(msg ...interface{})   { m.Called(msg...) }
func (m *loggerMock) Error(msg ...interface{})   { m.Called(msg...) }
func (m *loggerMock) Info(msg ...interface{})    { m.Called(msg...) }
func (m *loggerMock) Verbose(msg ...interface{}) { m.Called(msg...) }
func (m *loggerMock) Warning(msg ...interface{}) { m.Called(msg...) }

func ref[T any](t T) *T {
	return &t
}

var _ logging.LoggerInterface = (*loggerMock)(nil)
