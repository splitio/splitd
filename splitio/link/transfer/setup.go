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

func (c ConnType) String() string {
	switch c {
	case ConnTypeUnixSeqPacket:
		return "unix-seqpacket"
	case ConnTypeUnixStream:
		return "unix-stream"
	default:
		return "invalid-socket-type"
	}
}

const (
	ConnTypeUnixSeqPacket ConnType = 1
	ConnTypeUnixStream    ConnType = 2
)

var (
	ErrInvalidConnType = errors.New("invalid conn type")
)

func NewAcceptor(logger logging.LoggerInterface, o *Options, listenerConfig *AcceptorConfig) (*Acceptor, error) {

	var address net.Addr
	var ff FramingWrapperFactory
	switch o.ConnType {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.Address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.Address}
		ff = lpFramerFromConn
	default:
		return nil, ErrInvalidConnType
	}

	cf := func(c net.Conn) RawConn { return newConnWrapper(c, ff, o) }
	return newAcceptor(address, cf, logger, listenerConfig), nil
}

func NewClientConn(logger logging.LoggerInterface, o *Options) (RawConn, error) {

	var address net.Addr
	var ff FramingWrapperFactory
	switch o.ConnType {
	case ConnTypeUnixSeqPacket:
		address = &net.UnixAddr{Net: "unixpacket", Name: o.Address}
	case ConnTypeUnixStream:
		address = &net.UnixAddr{Net: "unix", Name: o.Address}
		ff = lpFramerFromConn
	default:
		return nil, ErrInvalidConnType
	}

	c, err := net.Dial(address.Network(), address.String())
	if err != nil {
		return nil, fmt.Errorf("error creating connection: %w", err)
	}

	return newConnWrapper(c, ff, o), nil
}

type Options struct {
	ConnType     ConnType
	Address      string
	BufferSize   int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func DefaultOpts() Options {
	return Options{
		ConnType:     ConnTypeUnixSeqPacket,
		Address:      "/var/run/splitd.sock",
		BufferSize:   1024,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
}

// helpers

func lpFramerFromConn(c net.Conn) framing.Interface { return framing.NewLengthPrefix(c) }
