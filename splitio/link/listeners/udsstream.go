package listeners

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/listeners/framing"
)

type UDSStream struct {
	lll    *net.UnixListener // low-level-listener
	fn     string
	logger logging.LoggerInterface
	framer framing.Interface
}

func newUDSStreamListener(opts *options) (*UDSStream, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: opts.fn, Net: "unix"})
	if err != nil {
		return nil, fmt.Errorf("error listening on provided address: %w", err)
	}
	return &UDSStream{
		lll:    l,
		fn:     opts.fn,
		logger: opts.logger,
		framer: &framing.LengthPrefixImpl{},
	}, nil
}

func (l *UDSStream) Listen(onClientAttached OnClientAttachedCallback) <-chan error {
	ret := make(chan error, 1)
	go func() {
		defer os.Remove(l.fn)
		for {
			conn, err := l.lll.AcceptUnix()
			if err != nil {
				var toSend error
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					toSend = err
				}
				ret <- toSend
				return
			}
			onClientAttached(&UnixStreamClientConn{
				listener: l,
				conn:     conn,
				framer:   l.framer,
			})
		}
	}()
	return ret
}

func (l *UDSStream) Shutdown() error {
	err := l.lll.Close()
	if err != nil {
		return fmt.Errorf("error closing UDSStream listener: %w", err)
	}
	return nil
}

type UnixStreamClientConn struct {
	listener *UDSStream
	conn     *net.UnixConn
	framer   framing.Interface
}

// ReceiveMessage implements ClientConnection
func (c *UnixStreamClientConn) ReceiveMessage() ([]byte, error) {
	var buf [bufSize]byte
	n, err := c.framer.ReadFrame(c.conn, buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, err
		}
		return nil, fmt.Errorf("error reading from UDSStream connection: %w", err)
	}

	if n == bufSize-1 {
		return nil, ErrBufferTooSmall
	}
	return buf[:n], nil
}

// SendMessage implements ClientConnection
func (c *UnixStreamClientConn) SendMessage(data []byte) error {
	data = c.framer.Frame(data)
	sent, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("error when sending message via UDSStream to client: %w", err)
	}

	if sent != len(data) {
		return fmt.Errorf("sent data mismatch. expected: %d, got: %d", len(data), sent)
	}
	return nil
}

// Shutdown implements ClientConnection
func (u *UnixStreamClientConn) Shutdown() error {
	err := u.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection")
	}
	return nil
}

var _ ClientConnection = (*UnixStreamClientConn)(nil)
var _ Raw = (*UDSStream)(nil)
