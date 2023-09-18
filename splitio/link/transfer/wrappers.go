package transfer

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/splitio/splitd/splitio/link/transfer/framing"
)

var (
	ErrSentDataMismatch = errors.New("sent data size mismatch")
	ErrBufferTooSmall   = errors.New("insufficient capacity in read buffer")
)

type FramingWrapperFactory func(c net.Conn) framing.Interface

type RawConn interface {
	ReceiveMessage() ([]byte, error)
	SendMessage(data []byte) error
	Shutdown() error
}

type BaseConn struct {
	conn         net.Conn
	readBuffer   []byte
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// Shutdown implements ClientConnection
func (b *BaseConn) Shutdown() error {
	err := b.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}
	return nil
}

func (b *BaseConn) setReadDeadline() error {
	if err := b.conn.SetReadDeadline(time.Now().Add(b.readTimeout)); err != nil {
		return fmt.Errorf("error setting read timeout: %w", err)
	}
	return nil
}

func (b *BaseConn) setWriteDeadline() error {
	if err := b.conn.SetWriteDeadline(time.Now().Add(b.writeTimeout)); err != nil {
		return fmt.Errorf("error setting write timeout: %w", err)
	}
	return nil
}

type PacketBasedConnection struct {
	BaseConn
}

// ReceiveMessage implements ClientConnection
func (c *PacketBasedConnection) ReceiveMessage() ([]byte, error) {
	if err := c.setReadDeadline(); err != nil {
		return nil, err
	}

	n, err := c.conn.Read(c.readBuffer)
	if err != nil {
		return nil, formatErrorIfApplicable("error reading from socket: %w", err)
	}

	if n == len(c.readBuffer) {
		return nil, ErrBufferTooSmall
	}
	return c.readBuffer[:n], nil
}

// SendMessage implements ClientConnection
func (c *PacketBasedConnection) SendMessage(data []byte) error {
	if err := c.setWriteDeadline(); err != nil {
		return err
	}

	_, err := c.conn.Write(data)
	if err != nil {
		return formatErrorIfApplicable("error when sending message to client: %w", err)
	}

	return nil
}

type StreamBasedConnection struct {
	BaseConn
	framer framing.Interface
}

func (c *StreamBasedConnection) ReceiveMessage() ([]byte, error) {
	if err := c.setReadDeadline(); err != nil {
		return nil, err
	}

	n, err := c.framer.ReadFrame(c.readBuffer)
	if err != nil {
		return nil, formatErrorIfApplicable("error reading frame: %w", err)
	}

	return c.readBuffer[:n], nil
}

func (c *StreamBasedConnection) SendMessage(data []byte) error {
	if err := c.setWriteDeadline(); err != nil {
		return err
	}

	if _, err := c.framer.WriteFrame(data); err != nil {
		return formatErrorIfApplicable("error writing frame: %w", err)
	}

	return nil
}

// io.EOF shuold NOT be wrapped since the go standard library queries by exact comparison and does
// not check the error chain (errors.Is), hence the need for a conditional wrapper
func formatErrorIfApplicable(message string, err error) error {

	if err == nil {
		return nil
	}

	if err == io.EOF {
		return err
	}

	return fmt.Errorf(message, err)
}

func newConnWrapper(c net.Conn, f FramingWrapperFactory, o *Options) RawConn {

	bc := BaseConn{
		conn:         c,
		readBuffer:   make([]byte, o.BufferSize),
		readTimeout:  o.ReadTimeout,
		writeTimeout: o.WriteTimeout,
	}

	if f == nil {
		return &PacketBasedConnection{BaseConn: bc}
	}
	return &StreamBasedConnection{framer: f(c), BaseConn: bc}
}

var _ RawConn = (*PacketBasedConnection)(nil)
var _ RawConn = (*StreamBasedConnection)(nil)
