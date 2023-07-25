package transfer

import (
	"io"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/transfer/framing"
	"github.com/stretchr/testify/assert"
)

func TestAcceptor(t *testing.T) {
	// This test sets up an acceptor with the following params:
	// - a queue size of 1
	// - a 100ms timeout for items in the waitqueue
	// - a 300ms delay in the handler (so that by the time the 2nd client tries to connect, the first one is busy)
	//
	// 2 clients will try to connect and do some write & reads.
	// First client will successfully connect and exchange some information
	// Second client's server-end of the socket will be closed after the timeout
	// The write not error out (though nothing is written), but will notice that the socket hasbeen remotely closed and update it's state
	// The following read will report an EOF

	logger := logging.NewLogger(nil)
	dir, err := os.MkdirTemp(os.TempDir(), "acceptortest")
	assert.Nil(t, err)

	serverSockFN := path.Join(dir, "acctest.sock")

	connOpts := DefaultOpts()

	acceptorConfig := DefaultAcceptorConfig()
	acceptorConfig.AcceptTimeout = 100 * time.Millisecond
	acceptorConfig.MaxSimultaneousConnections = 1
	acc := newAcceptor(&net.UnixAddr{Net: "unix", Name: serverSockFN}, func(c net.Conn) RawConn {
		return newConnWrapper(c, &framing.LengthPrefixImpl{}, &connOpts)
	}, logger, &acceptorConfig)

	endc, err := acc.Start(func(c RawConn) {
		message, err := c.ReceiveMessage()
		assert.Nil(t, err)
		assert.Equal(t, "some", string(message))
		assert.Nil(t, c.SendMessage([]byte("thing")))
		time.Sleep(300 * time.Millisecond)
	})
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, acc.Shutdown())
		assert.Nil(t, <-endc)
	}()

	time.Sleep(1 * time.Second) // to ensure server is started

	clientOpts := DefaultOpts()
	clientOpts.Address = serverSockFN
	clientOpts.ConnType = ConnTypeUnixStream

	client1, err := NewClientConn(&clientOpts)
	assert.Nil(t, err)
	assert.NotNil(t, client1)
	err = client1.SendMessage([]byte("some"))
	assert.Nil(t, err)
	recv, err := client1.ReceiveMessage()
	assert.Nil(t, err)
	assert.Equal(t, []byte("thing"), recv)

	client2, err := NewClientConn(&clientOpts)
	assert.Nil(t, err)
	err = client2.SendMessage([]byte("some"))
	assert.Nil(t, err) // write doesn't fail. instead causes the transition of the socket to EOF state
	recv, err = client1.ReceiveMessage()
	assert.Nil(t, recv)
	assert.ErrorIs(t, err, io.EOF)
}
