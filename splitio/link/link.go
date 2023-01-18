package link

import (
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/listeners"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/service"
	"github.com/splitio/splitd/splitio/sdk"
)

type Option func(*opts) error

func WithSockType(s string) Option {
	return func(o *opts) error {
		switch s {
		case "unix-seqpacket":
			o.sockType = listeners.ListenerTypeUnixSeqPacket
			return nil
		case "unix-stream":
			o.sockType = listeners.ListenerTypeUnixStream
			return nil
		}
		return fmt.Errorf("unknown listener type '%s'", s)
	}
}

func WithSockFN(s string) Option {
	return func(o *opts) error {
		o.sockFN = s
		return nil
	}
}

func WithSerialization(s string) Option {
	return func(o *opts) error {
		switch s {
		case "msgpack":
			o.serialization = serializer.MsgPack
			return nil
		}
		return fmt.Errorf("unknown serialization mechanism '%s'", s)
	}
}
func WithProtocol(p string) Option {
	return func(o *opts) error {
		switch p {
		case "v1":
			o.protocolV = protocol.V1
			return nil
		}
		return fmt.Errorf("unkown protocol version '%s'", p)
	}

}

type opts struct {
	sockType      listeners.ListenerType
	sockFN        string
	serialization serializer.Mechanism
	protocolV     protocol.Version
}

func (o *opts) populate(options []Option) error {
	for _, configure := range options {
		err := configure(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func defaultOpts() opts {
	return opts{
		sockType:      listeners.ListenerTypeUnixSeqPacket,
		sockFN:        "/var/run/splitd.sock",
		serialization: serializer.MsgPack,
		protocolV:     protocol.V1,
	}
}

func Listen(logger logging.LoggerInterface, sdkFacade sdk.Interface, os ...Option) (<-chan error, func() error, error) {

	opts := defaultOpts()
	err := opts.populate(os)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing config options: %w", err)
	}

	svc, err := service.New(logger, sdkFacade, opts.protocolV, opts.serialization)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up service handler: %w", err)
	}

	l, err := listeners.Create(opts.sockType, listeners.WithFileName(opts.sockFN))
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up listener: %w", err)
	}

	return l.Listen(svc.HandleNewClient), l.Shutdown, nil
}
