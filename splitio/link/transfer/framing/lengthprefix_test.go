package framing

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLengthPrefixRead(t *testing.T) {

    // the test will send the following string in 3 parts (one per word)
    // we expect 4 calls to read:
    // - 1: total size
    // - 2: "something"
    // - 3: " to"
    // - 4: " send"
    sock := &sockMock{}
    message := "SOMETHING TO SEND"
    messageSize := len(message)
    sock.On("Read", mock.Anything).
        Run(func(args mock.Arguments) { copy(args.Get(0).([]byte), encodeSize(messageSize, nil)) }).
        Return(4, (error)(nil)).Once()
    sock.
        On("Read", mock.Anything).
        Run(func(args mock.Arguments) { copy(args.Get(0).([]byte), []byte("SOMETHING")) }).
        Return(len("SOMETHING"), (error)(nil)).Once()
    sock.
        On("Read", mock.Anything).
        Run(func(args mock.Arguments) { copy(args.Get(0).([]byte), []byte(" TO")) }).
        Return(len(" TO"), (error)(nil)).Once()
    sock.
        On("Read", mock.Anything).
        Run(func(args mock.Arguments) { copy(args.Get(0).([]byte), []byte(" SEND")) }).
        Return(len(" SEND"), (error)(nil)).Once()


    var lp LengthPrefixImpl
    var buffer [2048]byte
    n, err := lp.ReadFrame(sock, buffer[:])
    assert.Nil(t, err)
    assert.Equal(t, "SOMETHING TO SEND", string(buffer[:n]))
}

func TestLengthPrefixWrite(t *testing.T) {

    // the test will write the following string in 3 parts
    // - 1: total size + "SOME"
    // - 2: "THING "
    // - 3: " TO SEND"
    sock := &sockMock{}
    message := "SOMETHING TO SEND"
    messageSize := len(message)
    sock.On("Write", append(encodeSize(messageSize, nil), []byte("SOMETHING TO SEND")...)).
        Return(8, (error)(nil)).Once()
    sock.
        On("Write", []byte("THING TO SEND")).
        Return(6, (error)(nil)).Once()
    sock.
        On("Write", []byte("TO SEND")).
        Return(7, (error)(nil)).Once()

    var lp LengthPrefixImpl
    n, err := lp.WriteFrame(sock, []byte("SOMETHING TO SEND"))
    assert.Nil(t, err)
    assert.Equal(t, len("SOMETHING TO SEND"), n)
}



type sockMock struct {
	mock.Mock
}

// Write implements io.Writer
func (s *sockMock) Write(p []byte) (n int, err error) {
	args := s.Called(p)
	return args.Int(0), args.Error(1)
}

// Read implements io.Reader
func (s *sockMock) Read(p []byte) (n int, err error) {
	args := s.Called(p)
	return args.Int(0), args.Error(1)
}

var _ io.Reader = (*sockMock)(nil)
var _ io.Writer = (*sockMock)(nil)
