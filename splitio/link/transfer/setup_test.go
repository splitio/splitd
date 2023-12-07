package transfer

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/stretchr/testify/assert"
)

func TestConnType(t *testing.T) {
	assert.Equal(t, "unix-seqpacket", ConnTypeUnixSeqPacket.String())
	assert.Equal(t, "unix-stream", ConnTypeUnixStream.String())
	assert.Equal(t, "invalid-socket-type", ConnType(123).String())
}

func TestEnsureAddressIsUsable(t *testing.T) {

	logger := logging.NewLogger(nil)
	assert.Nil(t, ensureAddressUsable(logger, &net.UDPAddr{}))
	assert.Nil(t, ensureAddressUsable(logger, &net.TCPAddr{}))
	assert.Nil(t, ensureAddressUsable(logger, &net.UnixAddr{Name: "/some/nonexistent/file"}))

	// test unknown error (in this case trying to connect to a different socket type)
	ready := make(chan struct{})
	path := filepath.Join(os.TempDir(), "splitd_test_ensure_address_usable.sock")
	os.Remove(path) // por las dudas
	go func() {
		l, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
		assert.Nil(t, err)
		defer l.Close()

		l.SetDeadline(time.Now().Add(1 * time.Second))
		go func() {
			time.Sleep(100 * time.Millisecond)
			ready <- struct{}{}
		}()
		l.Accept()
	}()
	<-ready
	assert.ErrorContains(t, ensureAddressUsable(logger, &net.UnixAddr{Name: path, Net: "unixpacket"}), "unknown error when testing for a dead socket")

	// test socket in use error
	ready = make(chan struct{})
	path = filepath.Join(os.TempDir(), "splitd_test_ensure_address_usable2.sock")
	os.Remove(path) // por las dudas
	go func() {
		l, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
		assert.Nil(t, err)
		defer l.Close()

		l.SetDeadline(time.Now().Add(1 * time.Second))
		go func() {
			time.Sleep(100 * time.Millisecond)
			ready <- struct{}{}
		}()
		l.Accept()
	}()
	<-ready
	assert.ErrorIs(t, ErrServiceAddressInUse, ensureAddressUsable(logger, &net.UnixAddr{Name: path, Net: "unix"}))
}
