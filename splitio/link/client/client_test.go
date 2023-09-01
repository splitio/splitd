package client

import (
	"os"
	"strconv"
	"testing"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/protocol"
	v1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	smocks "github.com/splitio/splitd/splitio/link/serializer/mocks"
	"github.com/splitio/splitd/splitio/link/transfer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClientConstruction(t *testing.T) {
	connMock := &mocks.RawConnMock{}
	connMock.On("SendMessage", mock.Anything).Once().Return(nil)
	connMock.On("ReceiveMessage").Once().Return([]byte("registerResp"), nil)
    connMock.On("Shutdown").Once().Return(nil)
	serializeMock := &smocks.SerializerMock{}
	serializeMock.On("Serialize", mock.Anything).Once().Return([]byte("serializedRegister"), nil)
	serializeMock.On("Parse", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
        args.Get(1).(*v1.ResponseWrapper[v1.RegisterPayload]).Status = v1.ResultOk
    }).Return(nil)
	logger := logging.NewLogger(nil)
	c, err := New(logger, connMock, serializeMock, Options{Protocol: protocol.V1, ImpressionsFeedback: false})
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Nil(t, c.Shutdown())
}

func TestClientUnknownProtocol(t *testing.T) {
	connMock := &mocks.RawConnMock{}
	serializeMock := &smocks.SerializerMock{}
	logger := logging.NewLogger(nil)
	c, err := New(logger, connMock, serializeMock, Options{Protocol: protocol.Version(252), ImpressionsFeedback: false})
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "unknown protocol")
}

func TestDefaultOpts(t *testing.T) {
    do := DefaultOptions()
    assert.Equal(t, strconv.Itoa(os.Getpid()), do.ID)
    assert.Equal(t, false, do.ImpressionsFeedback)
    assert.Equal(t, protocol.V1, do.Protocol)
}
