package transfer

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHappyPathNoFraming(t *testing.T) {
	conn := &connMock{}
    wrapped := newConnWrapper(conn, nil, &Options{BufferSize: 1024})

	recent := func(t time.Time) bool { return time.Now().Sub(t) < time.Millisecond }

	conn.On("SetWriteDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Write", []byte("SOME MESSAGE")).Return(len("SOME MESSAGE"), (error)(nil)).Once()
	conn.On("SetReadDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Read", mock.Anything).
		Run(func(args mock.Arguments) { copy(args.Get(0).([]byte), []byte("SOME MESSAGE")) }).
		Return(len("SOME MESSAGE"), (error)(nil)).
		Once()

	err := wrapped.SendMessage([]byte("SOME MESSAGE"))
	assert.Nil(t, err)

	message, err := wrapped.ReceiveMessage()
	assert.Nil(t, err)
	assert.Equal(t, []byte("SOME MESSAGE"), message)
}

func TestWriteTimeout(t *testing.T) {
	conn := &connMock{}
	wrapped := newConnWrapper(conn, nil, &Options{BufferSize: 1024})

	recent := func(t time.Time) bool { return time.Now().Sub(t) < time.Millisecond }

	conn.On("SetWriteDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Write", []byte("SOME MESSAGE")).Return(int(0), context.DeadlineExceeded).Once()

	err := wrapped.SendMessage([]byte("SOME MESSAGE"))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWriteSizeMismatch(t *testing.T) {
	conn := &connMock{}
    wrapped := newConnWrapper(conn, nil, &Options{BufferSize: 1024})

	recent := func(t time.Time) bool { return time.Now().Sub(t) < time.Millisecond }

	conn.On("SetWriteDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Write", []byte("SOME MESSAGE")).Return(int(5), (error)(nil)).Once()

	err := wrapped.SendMessage([]byte("SOME MESSAGE"))
	assert.ErrorIs(t, err, ErrSentDataMismatch)
}

func TestReadEOF(t *testing.T) {
	conn := &connMock{}
    wrapped := newConnWrapper(conn, nil, &Options{BufferSize: 1024})

	recent := func(t time.Time) bool { return time.Now().Sub(t) < time.Millisecond }

	conn.On("SetReadDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Read", mock.Anything).Return(0, io.EOF).Once()

	message, err := wrapped.ReceiveMessage()
	assert.Nil(t, message)
	assert.Equal(t, io.EOF, err)
}

func TestReadNonEOFError(t *testing.T) {
	conn := &connMock{}
    wrapped := newConnWrapper(conn, nil, &Options{BufferSize: 1024})

	recent := func(t time.Time) bool { return time.Now().Sub(t) < time.Millisecond }

    var someErr = errors.New("some")
	conn.On("SetReadDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Read", mock.Anything).Return(0, someErr).Once()

	message, err := wrapped.ReceiveMessage()
	assert.Nil(t, message)
	assert.NotEqual(t, err, someErr)
    assert.ErrorIs(t, err,  someErr)
}

func TestReadInsufficientBufferSpace(t *testing.T) {
	conn := &connMock{}
    wrapped := newConnWrapper(conn, nil, &Options{BufferSize: 1024})

	recent := func(t time.Time) bool { return time.Now().Sub(t) < time.Millisecond }

	conn.On("SetReadDeadline", mock.MatchedBy(recent)).Return((error)(nil)).Once()
	conn.On("Read", mock.Anything).Return(1024, (error)(nil)).Once()

	message, err := wrapped.ReceiveMessage()
	assert.Nil(t, message)
    assert.Equal(t, ErrBufferTooSmall, err)
}

type connMock struct {
	mock.Mock
}

// Close implements net.Conn
func (c *connMock) Close() error {
	args := c.Called()
	return args.Error(0)
}

// LocalAddr implements net.Conn
func (c *connMock) LocalAddr() net.Addr {
	args := c.Called()
	return args.Get(0).(net.Addr)

}

// Read implements net.Conn
func (c *connMock) Read(b []byte) (n int, err error) {
	args := c.Called(b)
	return args.Int(0), args.Error(1)

}

// RemoteAddr implements net.Conn
func (c *connMock) RemoteAddr() net.Addr {
	args := c.Called()
	return args.Get(0).(net.Addr)

}

// SetDeadline implements net.Conn
func (c *connMock) SetDeadline(t time.Time) error {
	args := c.Called(t)
	return args.Error(0)

}

// SetReadDeadline implements net.Conn
func (c *connMock) SetReadDeadline(t time.Time) error {
	args := c.Called(t)
	return args.Error(0)

}

// SetWriteDeadline implements net.Conn
func (c *connMock) SetWriteDeadline(t time.Time) error {
	args := c.Called(t)
	return args.Error(0)

}

// Write implements net.Conn
func (c *connMock) Write(b []byte) (n int, err error) {
	args := c.Called(b)
	return args.Int(0), args.Error(1)

}

var c net.Conn = (*connMock)(nil)
