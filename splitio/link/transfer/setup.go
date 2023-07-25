package transfer

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/transfer/framing"
)

type ConnType int

const (
	ConnTypeUnixSeqPacket ConnType = 1
	ConnTypeUnixStream    ConnType = 2
)

var (
	ErrInvalidConnType = errors.New("invalid listener type")
)

func NewAcceptor(logger logging.LoggerInterface, o *Options, listenerConfig *AcceptorConfig) (*Acceptor, error) {

	var address net.Addr
	var framer framing.Interface
	switch o.ConnType {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.Address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.Address}
		framer = &framing.LengthPrefixImpl{}
	default:
		return nil, ErrInvalidConnType
	}

	connFactory := func(c net.Conn) RawConn { return newConnWrapper(c, framer, o) }
	return newAcceptor(address, connFactory, logger, listenerConfig), nil
}

func NewClientConn(o *Options) (RawConn, error) {

	var address net.Addr
	var framer framing.Interface
	switch o.ConnType {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.Address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.Address}
		framer = &framing.LengthPrefixImpl{}
	default:
		return nil, ErrInvalidConnType
	}

	c, err := net.Dial(address.Network(), address.String())
	if err != nil {
		return nil, fmt.Errorf("error creating connection: %w", err)
	}

	return newConnWrapper(c, framer, o), nil
}

type Option func(*Options)

/*
func WithAddress(address string) Option                { return func(o *Options) { o.Address = address } }
func WithType(t ConnType) Option                       { return func(o *Options) { o.ConnType = t } }
func WithLogger(logger logging.LoggerInterface) Option { return func(o *Options) { o.Logger = logger } }
func WithBufSize(s int) Option                         { return func(o *Options) { o.BufferSize = s } }
func WithMaxConns(m int) Option                        { return func(o *Options) { o.MaxSimultaneousConnections = m } }
func WithReadTimeout(d time.Duration) Option           { return func(o *Options) { o.ReadTimeout = d } }
func WithWriteTimeout(d time.Duration) Option          { return func(o *Options) { o.WriteTimeout = d } }
func WithAcceptTimeout(d time.Duration) Option         { return func(o *Options) { o.AcceptTimeout = d } }
*/

type Options struct {
	ConnType                   ConnType
	Address                    string
	Logger                     logging.LoggerInterface
	BufferSize                 int
	ReadTimeout                time.Duration
	WriteTimeout               time.Duration
}
/*
func (o *Options) Parse(opts []Option) {
	for _, configure := range opts {
		configure(o)
	}
}
*/
func DefaultOpts() Options {
	return Options{
		ConnType:                   ConnTypeUnixSeqPacket,
		Address:                    "/var/run/splitd.sock",
		Logger:                     logging.NewLogger(nil),
		BufferSize:                 1024,
		ReadTimeout:                1 * time.Second,
		WriteTimeout:               1 * time.Second,
	}
}
