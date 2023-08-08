package v1

import (
	"errors"
	"io"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
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

func TestManagePanicRecovers(t *testing.T) {
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Panic("some panic")
	rawConnMock.On("Shutdown", mock.Anything).Return(nil)

	logger := &loggerMock{}
	logger.On("Error", "CRITICAL - connection handler is panicking: ", "some panic").Once()

	serializerMock := &serializerMocks.SerializerMock{}
	sdkMock := &sdkMocks.SDKMock{}

	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	cm.Manage()
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)

	logger.AssertExpectations(t)
}

func TestFetchRPC(t *testing.T) {
	// error reading from conn
	someErr := errors.New("someConnErr")
	rawConnMock := &transferMocks.RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), someErr)
	rawConnMock.On("Shutdown", mock.Anything).Return(nil)
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
	rawConnMock.On("Shutdown", mock.Anything).Return(nil)
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
	rawConnMock.On("Shutdown", mock.Anything).Return(nil)
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
	rawConnMock.On("Shutdown", mock.Anything).Return(nil)
	serializerMock = &serializerMocks.SerializerMock{}
	serializerMock.On("Serialize", mock.Anything).Return([]byte{}, nil)
	cm = NewClientManager(rawConnMock, logger, nil, serializerMock)
	err = cm.sendResponse(nil)
	assert.ErrorContains(t, err, "someConnErr")
}

func TestHandleRPCErrors(t *testing.T) {
	logger := logging.NewLogger(nil)
	cm := NewClientManager(nil, logger, nil, nil)
	res, err := cm.handleRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCTreatment})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "first call must be 'register'")

	// register wrong args
	res, err = cm.handleRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCRegister, Args: []interface{}{1, "hola"}})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "error parsing register arguments")

    // set the config to allow other rpcs to be handled
    cm.clientConfig = &types.ClientConfig{ReturnImpressionData: true}

	// treatment wrong args
	res, err = cm.handleRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCTreatment, Args: []interface{}{1, "hola"}})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "error parsing treatment arguments")

	// register wrong args
	res, err = cm.handleRPC(&v1.RPC{RPCBase: protocol.RPCBase{Version: protocol.V1}, OpCode: v1.OCTreatments, Args: []interface{}{1, "hola"}})
	assert.Nil(t, res)
	assert.ErrorContains(t, err, "error parsing treatments arguments")
}

type loggerMock struct{ mock.Mock }

func (m *loggerMock) Debug(msg ...interface{})   { m.Called(msg...) }
func (m *loggerMock) Error(msg ...interface{})   { m.Called(msg...) }
func (m *loggerMock) Info(msg ...interface{})    { m.Called(msg...) }
func (m *loggerMock) Verbose(msg ...interface{}) { m.Called(msg...) }
func (m *loggerMock) Warning(msg ...interface{}) { m.Called(msg...) }

var _ logging.LoggerInterface = (*loggerMock)(nil)
