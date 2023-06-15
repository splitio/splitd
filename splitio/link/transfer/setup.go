package transfer

import (
	"errors"
	"fmt"
	"net"

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

func NewAcceptor(opts ...Option) (*Acceptor, error) {

	o := defaultOpts()
	for _, configure := range opts {
		configure(&o)
	}

	var address net.Addr
	var framer framing.Interface
	switch o.ctype {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.address}
		framer = &framing.LengthPrefixImpl{}
	default:
		return nil, ErrInvalidConnType
	}

	connFactory := func(c net.Conn) RawConn { return newConnWrapper(c, framer, o.bufsize) }
	return newAcceptor(address, connFactory, o.logger, o.maxSimultaneousConnections), nil
}

func NewClientConn(opts ...Option) (RawConn, error) {
	o := defaultOpts()
	for _, configure := range opts {
		configure(&o)
	}

	var address net.Addr
	var framer framing.Interface
	switch o.ctype {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.address}
		framer = &framing.LengthPrefixImpl{}
	default:
		return nil, ErrInvalidConnType
	}

	c, err := net.Dial(address.Network(), address.String())
	if err != nil {
		return nil, fmt.Errorf("error creating connection: %w", err)
	}

	return newConnWrapper(c, framer, o.bufsize), nil
}

type Option func(*options)

func WithAddress(address string) Option                { return func(o *options) { o.address = address } }
func WithType(t ConnType) Option                       { return func(o *options) { o.ctype = t } }
func WithLogger(logger logging.LoggerInterface) Option { return func(o *options) { o.logger = logger } }
func WithBufSize(s int) Option                         { return func(o *options) { o.bufsize = s } }
func WithMaxConns(m int) Option                        { return func(o *options) { o.maxSimultaneousConnections = m } }

type options struct {
	ctype                      ConnType
	address                    string
	logger                     logging.LoggerInterface
	bufsize                    int
	maxSimultaneousConnections int
}

func defaultOpts() options {
	return options{
		ctype:   ConnTypeUnixSeqPacket,
		address: "/var/run/splitd.sock",
		logger:  logging.NewLogger(nil),
		bufsize: 1024,
	}
}
