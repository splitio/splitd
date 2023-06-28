package link

import (
	"fmt"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/client"
	"github.com/splitio/splitd/splitio/link/common"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/service"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"
)

func Listen(logger logging.LoggerInterface, sdkFacade sdk.Interface, os ...Option) (<-chan error, func() error, error) {

	var opts Options
	err := opts.populate(os)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing config options: %w", err)
	}

	acceptor, err := transfer.NewAcceptor(opts.forTransfer()...)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up transfer module: %w", err)
	}

	svc, err := service.New(logger, sdkFacade, opts.forApp()...)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up service handler: %w", err)
	}

	ec, err := acceptor.Start(svc.HandleNewClient)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up listener: %w", err)
	}

	return ec, acceptor.Shutdown, nil
}

func Consumer(logger logging.LoggerInterface, os ...Option) (client.Interface, error) {

	var opts Options
	err := opts.populate(os)
	if err != nil {
		return nil, fmt.Errorf("error parsing config options: %w", err)
	}

	conn, err := transfer.NewClientConn(opts.forTransfer()...)
	if err != nil {
		return nil, fmt.Errorf("errpr creating connection: %w", err)
	}

	return client.New(logger, conn, opts.forApp()...)
}

type Option func(*Options) error

func WithSockType(s string) Option {
	return func(o *Options) error {
		switch s {
		case "unix-seqpacket":
			o.sockType = transfer.ConnTypeUnixSeqPacket
			return nil
		case "unix-stream":
			o.sockType = transfer.ConnTypeUnixStream
			return nil
		}
		return fmt.Errorf("unknown listener type '%s'", s)
	}
}

func WithAddress(s string) Option {
	return func(o *Options) error { o.address = s; return nil }
}

func WithBufSize(b int) Option {
	return func(o *Options) error { o.bufSize = b; return nil }
}

func WithSerialization(s string) Option {
	return func(o *Options) error {
		switch s {
		case "msgpack":
			o.serialization = serializer.MsgPack
			return nil
		}
		return fmt.Errorf("unknown serialization mechanism '%s'", s)
	}
}

func WithProtocol(p string) Option {
	return func(o *Options) error {
		switch p {
		case "v1":
			o.protocolV = protocol.V1
			return nil
		}
		return fmt.Errorf("unkown protocol version '%s'", p)
	}

}

func WithMaxSimultaneousConns(m int) Option {
	return func(o *Options) error { o.maxSimulateneousConns = m; return nil }
}

func WithReadTimeoutMs(m int) Option {
	return func(o *Options) error { o.readTimeoutMS = m; return nil }
}

func WithWriteTimeoutMs(m int) Option {
	return func(o *Options) error { o.writeTimeoutMS = m; return nil }
}

func WithAcceptTimeoutMs(m int) Option {
	return func(o *Options) error { o.acceptTimeoutMS = m; return nil }
}

type Options struct {
	sockType              transfer.ConnType
	address               string
	serialization         serializer.Mechanism
	protocolV             protocol.Version
	bufSize               int
	maxSimulateneousConns int
	readTimeoutMS         int
	writeTimeoutMS        int
	acceptTimeoutMS       int
}

func (o *Options) populate(options []Option) error {
	for _, configure := range options {
		err := configure(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) forTransfer() []transfer.Option {
	var toRet []transfer.Option
	if o.sockType != 0 {
		toRet = append(toRet, transfer.WithType(o.sockType))
	}
	if o.address != "" {
		toRet = append(toRet, transfer.WithAddress(o.address))
	}
	if o.bufSize != 0 {
		toRet = append(toRet, transfer.WithBufSize(o.bufSize))
	}
	if o.maxSimulateneousConns != 0 {
		toRet = append(toRet, transfer.WithMaxConns(o.maxSimulateneousConns))
	}
	if o.readTimeoutMS != 0 {
		toRet = append(toRet, transfer.WithReadTimeout(time.Duration(o.readTimeoutMS)*time.Millisecond))
	}
	if o.writeTimeoutMS != 0 {
		toRet = append(toRet, transfer.WithWriteTimeout(time.Duration(o.writeTimeoutMS)*time.Millisecond))
	}
	if o.acceptTimeoutMS != 0 {
		toRet = append(toRet, transfer.WithAcceptTimeout(time.Duration(o.acceptTimeoutMS)*time.Millisecond))
	}
	return toRet
}

func (o *Options) forApp() []common.Option {
	var toRet []common.Option
	if o.protocolV != 0 {
		toRet = append(toRet, common.WithProtocolV(o.protocolV))
	}
	if o.serialization != 0 {
		toRet = append(toRet, common.WithSerialization(o.serialization))
	}
	return toRet
}
