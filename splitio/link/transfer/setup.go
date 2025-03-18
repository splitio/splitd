package transfer

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
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
	ErrInvalidConnType     = errors.New("invalid conn type")
	ErrServiceAddressInUse = errors.New("provided socket file / address is already in use")
)

func NewAcceptor(forking bool, logger logging.LoggerInterface, o *Options, listenerConfig *AcceptorConfig) (*Acceptor, error) {

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

	if err := ensureAddressUsable(logger, address); err != nil {
		return nil, err
	}

	cf := func(c net.Conn) RawConn { return newConnWrapper(c, ff, o) }
	return newAcceptor(forking, address, cf, logger, listenerConfig), nil
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

func ensureAddressUsable(logger logging.LoggerInterface, address net.Addr) error {
	switch address.Network() {
	case "unix", "unixpacket":
		if _, err := os.Stat(address.String()); errors.Is(err, os.ErrNotExist) {
			return nil // file doesn't exist, we're ok
		}

		logger.Warning("The socket file exists. Testing if it's currently accepting connections")
		c, err := net.Dial(address.Network(), address.String())
		if err == nil {
			c.Close()
			return ErrServiceAddressInUse
		}

		logger.Warning("The socket appears to be from a previous (dead) execution. Will try to remove it")

		if !errors.Is(err, syscall.ECONNREFUSED) {
			return fmt.Errorf("unknown error when testing for a dead socket: %w", err)
		}

		// the socket seems to be bound to a dead process, will try removing it
		// so that a listener can be created
		if err := os.Remove(address.String()); err != nil {
			return fmt.Errorf("error removing dead-socket file from a previous execution: %w", err)
		}

		logger.Warning("Dead socket file removed successfuly")

	}
	return nil
}
