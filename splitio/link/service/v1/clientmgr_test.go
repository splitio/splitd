package v1

import (
	"errors"
	"io"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/protocol"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterAndTreatmentHappyPath(t *testing.T) {
	rawConnMock := &RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

	serializerMock := &SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(0)},
		}
	}).Once()
	serializerMock.On("Serialize", newRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatment,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", newTreatmentResp(true, "on", nil)).Return([]byte("successPayload"), nil).Once()

	sdkMock := &SDKMock{}
	sdkMock.
		On("Treatment", &types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3"}, "key", (*string)(nil), "someFeature", map[string]interface{}(nil)).
		Return("on", (*dtos.Impression)(nil), nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestRegisterWithImpsAndTreatmentHappyPath(t *testing.T) {
	rawConnMock := &RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("registrationMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successRegistration")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("SendMessage", []byte("successPayload")).Return(nil).Once()
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), io.EOF).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

	serializerMock := &SerializerMock{}
	serializerMock.On("Parse", []byte("registrationMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCRegister,
			Args:    []interface{}{"someID", "some_sdk-1.2.3", uint64(v1.RegisterFlagReturnImpressionData)},
		}
	}).Once()
	serializerMock.On("Serialize", newRegisterResp(true)).Return([]byte("successRegistration"), nil).Once()
	serializerMock.On("Parse", []byte("treatmentMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatment,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()
	serializerMock.On("Serialize", newTreatmentResp(true, "on", &v1.ListenerExtraData{Label: "l1", Timestamp: 1234556})).
		Return([]byte("successPayload"), nil).Once()

	sdkMock := &SDKMock{}
	sdkMock.
		On("Treatment",
			&types.ClientMetadata{ID: "someID", SdkVersion: "some_sdk-1.2.3", ReturnImpressionData: true},
			"key", (*string)(nil), "someFeature", map[string]interface{}(nil)).
		Return("on", &dtos.Impression{Label: "l1", Time: 1234556}, nil).Once()

	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Nil(t, err)
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestTreatmentWithoutRegister(t *testing.T) {
	rawConnMock := &RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte("treatmentMessage"), nil).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

	serializerMock := &SerializerMock{}
	serializerMock.On("Parse", []byte("treatmentMessage"), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		*args.Get(1).(*v1.RPC) = v1.RPC{
			RPCBase: protocol.RPCBase{Version: protocol.V1},
			OpCode:  v1.OCTreatment,
			Args:    []interface{}{"key", nil, "someFeature", map[string]interface{}(nil)},
		}
	}).Once()

	sdkMock := &SDKMock{}
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Contains(t, err.Error(), "first call must be 'register'")
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func TestConnectionFailureWhenReading(t *testing.T) {
	rawConnMock := &RawConnMock{}
	rawConnMock.On("ReceiveMessage").Return([]byte(nil), errors.New("something")).Once()
	rawConnMock.On("Shutdown").Return(nil).Once()

	serializerMock := &SerializerMock{}
	sdkMock := &SDKMock{}
	logger := logging.NewLogger(nil)
	cm := NewClientManager(rawConnMock, logger, sdkMock, serializerMock)
	err := cm.handleClientInteractions()
	assert.Contains(t, err.Error(), "error reading from conn")
	rawConnMock.AssertNumberOfCalls(t, "Shutdown", 1)
}

func newRegisterResp(ok bool) *v1.ResponseWrapper[v1.RegisterPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.RegisterPayload]{
		Status:  res,
		Payload: v1.RegisterPayload{},
	}
}

func newTreatmentResp(ok bool, treatment string, ilData *v1.ListenerExtraData) *v1.ResponseWrapper[v1.TreatmentPayload] {
	res := v1.ResultOk
	if !ok {
		res = v1.ResultInternalError
	}
	return &v1.ResponseWrapper[v1.TreatmentPayload]{
		Status:  res,
		Payload: v1.TreatmentPayload{
            Treatment: treatment,
            ListenerData: ilData,
        },
	}
}

// mocks

type RawConnMock struct {
	mock.Mock
}

// ReceiveMessage implements transfer.RawConn
func (m *RawConnMock) ReceiveMessage() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// SendMessage implements transfer.RawConn
func (m *RawConnMock) SendMessage(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

// Shutdown implements transfer.RawConn
func (m *RawConnMock) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

type SerializerMock struct {
	mock.Mock
}

// Parse implements serializer.Interface
func (m *SerializerMock) Parse(data []byte, v interface{}) error {
	args := m.Called(data, v)
	return args.Error(0)
}

// Serialize implements serializer.Interface
func (m *SerializerMock) Serialize(v interface{}) ([]byte, error) {
	args := m.Called(v)
	return args.Get(0).([]byte), args.Error(1)
}

type SDKMock struct {
	mock.Mock
}

// Treatment implements sdk.Interface
func (m *SDKMock) Treatment(md *types.ClientMetadata, key string, bucketingKey *string, feature string, attributes map[string]interface{}) (string, *dtos.Impression, error) {
	args := m.Called(md, key, bucketingKey, feature, attributes)
	return args.String(0), args.Get(1).(*dtos.Impression), nil
}

var _ transfer.RawConn = (*RawConnMock)(nil)
var _ serializer.Interface = (*SerializerMock)(nil)
var _ sdk.Interface = (*SDKMock)(nil)
