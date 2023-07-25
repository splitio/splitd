package v1

import (
	"errors"
	"io"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/protocol"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	serializerMocks "github.com/splitio/splitd/splitio/link/serializer/mocks"
	transferMocks "github.com/splitio/splitd/splitio/link/transfer/mocks"
	proto1Mocks "github.com/splitio/splitd/splitio/link/protocol/v1/mocks"
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
	rawConnMock.On("Shutdown").Return(nil).Once()

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
		Return(&sdk.Result{Treatment: "on"}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestRegisterAndTreatmentsHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentsMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

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
	serializerMock.On("Serialize", proto1Mocks.NewTreatmentsResp(true, []sdk.Result{
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
		).Return(map[string]sdk.Result{
		"feat1": {Treatment: "on"},
		"feat2": {Treatment: "off"},
		"feat3": {Treatment: "control"},
	}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestRegisterWithImpsAndTreatmentHappyPath(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

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
		Return(&sdk.Result{Treatment: "on", Impression: &dtos.Impression{Label: "l1", Time: 1234556}}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestTreatmentWithoutRegister(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

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
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestConnectionFailureWhenReading(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), errors.New("something")).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

	serializerMock := &serializerMocks.SerializerMock{}
	sdkMock := &sdkMocks.SDKMock{}
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Contains(t, err.Error(), "error reading from conn")
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}


