package transfer

import (
	"errors"
	"net"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/listeners/framing"
)

type ConnType int

const (
	ConnTypeUnixSeqPacket ConnType = 1
	ConnTypeUnixStream    ConnType = 2
)

var (
	ErrInvalidConnType = errors.New("invalid listener type")
)

type Option func(*options)

func WithAddress(address string) Option                { return func(o *options) { o.address = address } }
func WithType(t ConnType) Option                       { return func(o *options) { o.ctype = t } }
func WithLogger(logger logging.LoggerInterface) Option { return func(o *options) { o.logger = logger } }

func Create(lt ConnType, opts ...Option) (*Acceptor, error) {

	var o options
	for _, apply := range opts {
		apply(&o)
	}

	if o.logger == nil {
		o.logger = logging.NewLogger(nil)
	}

	var address net.Addr
	var framer framing.Interface
	switch lt {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.address}
		framer = &framing.LengthPrefixImpl{}
	default:
		return nil, ErrInvalidConnType
	}

	return NewAcceptor(
		address,
		nil,
		func(c net.Conn) RawConn { return newConnWrapper(c, framer, o.bufsize) },
		o.logger,
	), nil
}

type options struct {
	ctype   ConnType
	address string
	logger  logging.LoggerInterface
	bufsize int
}
