package listeners

import (
	"fmt"
	"net"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/protocol"
)

const (
	bufSize int = 1024
)

type UDSSeqPacket struct {
	lll    *net.UnixListener // low-level-listener
	fn     string
	logger logging.LoggerInterface
}

func newUDSSeqPacketListener(opts *options) (*UDSSeqPacket, error) {
	l, err := net.ListenUnix("unixpacket", &net.UnixAddr{Name: opts.fn, Net: "unixpacket"})
	if err != nil {
		return nil, fmt.Errorf("error listening on provided address: %w", err)
	}
	return &UDSSeqPacket{lll: l}, nil
}

func (l *UDSSeqPacket) Listen(onMessage OnRawMessageCallback) <-chan error {
	ret := make(chan error, 1)
	go func() {
		defer os.Remove(l.fn)
		for {
			conn, err := l.lll.AcceptUnix()
			if err != nil {
				ret <- err
				return
			}
			go newConnHandler(onMessage, l, conn).readAndHandle()
		}
	}()
	return ret
}

type connHandler struct {
	onMessage OnRawMessageCallback
	listener  *UDSSeqPacket
	conn      *net.UnixConn
	logger    logging.LoggerInterface
}

func newConnHandler(onMessage OnRawMessageCallback, listener *UDSSeqPacket, conn *net.UnixConn) *connHandler {
	return &connHandler{
		onMessage: onMessage,
		listener:  listener,
		conn:      conn,
	}
}

func (h *connHandler) readAndHandle() {
	var buf [bufSize]byte
	defer h.conn.Close()
	for {
		n, err := h.conn.Read(buf[:])
		if err != nil {
			// TODO(mredolatti): handle proeprly
			h.logger.Error("error reading from conn: ", err.Error())
			continue
		}

		if n == bufSize-1 {
			// TODO(mredolatti): handle proeprly
			h.logger.Error("insufficient buffer space to read whole message. discarding")
			continue
		}

		response, err := h.onMessage(buf[:n])
		if err != nil {
			if withResponse, ok := err.(protocol.ErrorWithResponse); ok {
				h.conn.Write(withResponse.ToResponse())
				continue
			} 
			h.logger.Error(fmt.Sprintf("fatal error con connection for metadata = TODO: %s. Aborting", err))
			// TODO(mredolatti): remove this
			panic(err.Error())
		}

		sent, err := h.conn.Write(response)
		if err != nil {
			h.logger.Error("error writing to conn: ", err.Error())
			h.conn.Close()
			return
		}

		if sent != n {
			h.logger.Error("failed to send all data. closing connection")
			h.conn.Close()
			return
		}
	}
}

/*
	l, err := net.ListenUnix("unixpacket", &net.UnixAddr{
		Name: "/tmp/unixdomain",
		Net:  "unixpacket",
	})
	if err != nil {
		panic(err)
	}
	defer os.Remove("/tmp/unixdomain")

}

func connHandler(conn *net.UnixConn) {
	for {
		var buf [1024]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			panic(err)
		}
		if n == 0 {
			continue
		}

		var msg common.Message
		err = msgpack.Unmarshal(buf[:n], &msg)
		if err != nil {
			panic(err)
		}

		fmt.Println("Func: ", msg.Func)
		fmt.Print("args: ")
		for _, arg := range msg.Args {
			fmt.Printf("%T::%+v,", arg, arg)
		}
		if _, err := conn.Write([]byte(msg.Func + "::" + "ok")); err != nil {
			panic(err)
		}
	}
	conn.Close()
}
*/

var _ Raw = (*UDSSeqPacket)(nil)
