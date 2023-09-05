package link

import (
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/client"
	"github.com/splitio/splitd/splitio/link/client/types"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/service"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
)

func Listen(logger logging.LoggerInterface, sdkFacade sdk.Interface, opts *ListenerOptions) (<-chan error, func() error, error) {

	acceptor, err := transfer.NewAcceptor(logger, &opts.Transfer, &opts.Acceptor)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up transfer module: %w", err)
	}

	s, err := serializer.Setup(opts.Serialization)
	if err != nil {
		return nil, nil, fmt.Errorf("error building serializer")
	}

	svc, err := service.New(logger, sdkFacade, s, opts.Protocol)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up service handler: %w", err)
	}

	ec, err := acceptor.Start(svc.HandleNewClient)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up listener: %w", err)
	}

	return ec, acceptor.Shutdown, nil
}

func Consumer(logger logging.LoggerInterface, opts *ConsumerOptions) (types.ClientInterface, error) {

	s, err := serializer.Setup(opts.Serialization)
	if err != nil {
		return nil, fmt.Errorf("error building serializer")
	}

	conn, err := transfer.NewClientConn(logger, &opts.Transfer)
	if err != nil {
		return nil, fmt.Errorf("errpr creating connection: %w", err)
	}

	return client.New(logger, conn, s, opts.Consumer)
}

type ListenerOptions struct {
	Transfer      transfer.Options
	Acceptor      transfer.AcceptorConfig
	Serialization serializer.Mechanism
	Protocol      protocol.Version
}

func DefaultListenerOptions() ListenerOptions {
	return ListenerOptions{
		Transfer:      transfer.DefaultOpts(),
		Acceptor:      transfer.DefaultAcceptorConfig(),
		Serialization: serializer.MsgPack,
		Protocol:      protocol.V1,
	}
}

type ConsumerOptions struct {
	Transfer      transfer.Options
	Consumer      client.Options
	Serialization serializer.Mechanism
}

func DefaultConsumerOptions() ConsumerOptions {
	return ConsumerOptions{
		Transfer:      transfer.DefaultOpts(),
		Consumer:      client.DefaultOptions(),
		Serialization: serializer.MsgPack,
	}
}
