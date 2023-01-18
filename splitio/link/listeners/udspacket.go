package listeners

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
)

type UDSSeqPacket struct {
	lll              *net.UnixListener // low-level-listener
	fn               string
	logger           logging.LoggerInterface
}

func newUDSSeqPacketListener(opts *options) (*UDSSeqPacket, error) {
	l, err := net.ListenUnix("unixpacket", &net.UnixAddr{Name: opts.fn, Net: "unixpacket"})
	if err != nil {
		return nil, fmt.Errorf("error listening on provided address: %w", err)
	}
	return &UDSSeqPacket{
		lll:    l,
		fn:     opts.fn,
		logger: opts.logger,
	}, nil
}

func (l *UDSSeqPacket) Listen(onClientAttached OnClientAttachedCallback) <-chan error {
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
			onClientAttached(&UnixPacketClientConn{
				listener: l,
				conn:     conn,
			})
		}
	}()
	return ret
}

func (l *UDSSeqPacket) Shutdown() error {
	err := l.lll.Close()
	if err != nil {
		return fmt.Errorf("error closing uds-seqpacket listener: %w", err)
	}
	return nil
}

type UnixPacketClientConn struct {
	listener *UDSSeqPacket
	conn     *net.UnixConn
}

// ReceiveMessage implements ClientConnection
func (c *UnixPacketClientConn) ReceiveMessage() ([]byte, error) {
	var buf [bufSize]byte
	n, err := c.conn.Read(buf[:])
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("error reading from UDSPacket connection: %w", err)
	}

	if n == bufSize-1 {
		return nil, ErrBufferTooSmall
	}
	return buf[:n], nil
}

// SendMessage implements ClientConnection
func (c *UnixPacketClientConn) SendMessage(data []byte) error {
	sent, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("error when sending message to client: %w", err)
	}

	if sent != len(data) {
		return ErrSentDataMismatch
	}
	return nil
}

// Shutdown implements ClientConnection
func (u *UnixPacketClientConn) Shutdown() error {
	err := u.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection")
	}
	return nil
}

var _ ClientConnection = (*UnixPacketClientConn)(nil)
var _ Raw = (*UDSSeqPacket)(nil)
