package link

import (
	"testing"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/client"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk/mocks"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	assert.Equal(t, ListenerOptions{
		Transfer:      transfer.DefaultOpts(),
		Acceptor:      transfer.DefaultAcceptorConfig(),
		Serialization: serializer.MsgPack,
		Protocol:      protocol.V1,
	},
		DefaultListenerOptions())

	assert.Equal(t, ConsumerOptions{
		Transfer:      transfer.DefaultOpts(),
		Consumer:      client.DefaultOptions(),
		Serialization: serializer.MsgPack,
	}, DefaultConsumerOptions())
}

func TestListenErrors(t *testing.T) {
	lo := DefaultListenerOptions()
	lo.Transfer.ConnType = transfer.ConnType(222)
	acc, shutdown, err := Listen(logging.NewLogger(nil), &mocks.SDKMock{}, &lo)
	assert.Nil(t, acc)
	assert.Nil(t, shutdown)
	assert.ErrorContains(t, err, "invalid conn type")

	lo = DefaultListenerOptions()
	lo.Protocol = protocol.Version(123)
	acc, shutdown, err = Listen(logging.NewLogger(nil), &mocks.SDKMock{}, &lo)
	assert.Nil(t, acc)
	assert.Nil(t, shutdown)
	assert.ErrorContains(t, err, "protocol")

	lo = DefaultListenerOptions()
	lo.Serialization = serializer.Mechanism(123)
	acc, shutdown, err = Listen(logging.NewLogger(nil), &mocks.SDKMock{}, &lo)
	assert.Nil(t, acc)
	assert.Nil(t, shutdown)
	assert.ErrorContains(t, err, "serializer")
}

func TestConsumerErrors(t *testing.T) {
	co := DefaultConsumerOptions()
    co.Transfer.ConnType = transfer.ConnType(123)
	client, err := Consumer(logging.NewLogger(nil), &co)
	assert.Nil(t, client)
	assert.ErrorContains(t, err, "invalid conn type")

	co = DefaultConsumerOptions()
	co.Serialization = serializer.Mechanism(123)
	client, err = Consumer(logging.NewLogger(nil), &co)
	assert.Nil(t, client)
	assert.ErrorContains(t, err, "serializer")

}
