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
	ErrSentDataMismatch    = errors.New("sent data size mismatch")
	ErrBufferTooSmall      = errors.New("insufficient capacity in read buffer")
)


type RawConn interface {
	ReceiveMessage() ([]byte, error)
	SendMessage(data []byte) error
	Shutdown() error
}

type Impl struct {
	conn       net.Conn
	readBuffer []byte
}

func newConnWrapper(c net.Conn, f framing.Interface, bufSize int) RawConn {
	if f != nil {
		c = &FramingRawConnWrapper{f: f, c: c}
	}

	return &Impl{
		conn:       c,
		readBuffer: make([]byte, bufSize),
	}
}

// ReceiveMessage implements ClientConnection
func (c *Impl) ReceiveMessage() ([]byte, error) {
	n, err := c.conn.Read(c.readBuffer)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("error reading from UDSPacket connection: %w", err)
	}

	if n == len(c.readBuffer)-1 {
		return nil, ErrBufferTooSmall
	}
	return c.readBuffer[:n], nil
}

// SendMessage implements ClientConnection
func (c *Impl) SendMessage(data []byte) error {
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
func (u *Impl) Shutdown() error {
	err := u.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection")
	}
	return nil
}

type FramingRawConnWrapper struct {
	f framing.Interface
	c net.Conn
}

func (w *FramingRawConnWrapper) Close() error                       { return w.c.Close() }
func (w *FramingRawConnWrapper) LocalAddr() net.Addr                { return w.c.LocalAddr() }
func (w *FramingRawConnWrapper) RemoteAddr() net.Addr               { return w.c.RemoteAddr() }
func (w *FramingRawConnWrapper) SetDeadline(t time.Time) error      { return w.c.SetDeadline(t) }
func (w *FramingRawConnWrapper) SetReadDeadline(t time.Time) error  { return w.c.SetReadDeadline(t) }
func (w *FramingRawConnWrapper) SetWriteDeadline(t time.Time) error { return w.c.SetWriteDeadline(t) }
func (w *FramingRawConnWrapper) Read(b []byte) (n int, err error)   { return w.f.ReadFrame(w.c, b) }
func (w *FramingRawConnWrapper) Write(b []byte) (n int, err error)  { return w.f.WriteFrame(w.c, b) }

var _ net.Conn = (*FramingRawConnWrapper)(nil)

var _ RawConn = (*Impl)(nil)
