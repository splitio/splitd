package consumer

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	netUnixPacket = "unixpacket"
)

type UDSPacketClient struct {
	conn *net.UnixConn
	fn  string
}

func newUDSPacketClient(opts *options) (*UDSPacketClient, error) {
	fn := makeSocketFile(opts.fn)
	laddr := net.UnixAddr{Name: fn, Net: netUnixPacket}
	conn, err := net.DialUnix(netUnixPacket, &laddr, &net.UnixAddr{
		Name: opts.fn,
		Net:  netUnixPacket,
	})

	if err != nil {
		os.Remove(fn)
		return nil, fmt.Errorf("error establishing conneciton: %w", err)
	}

	return &UDSPacketClient{
		conn: conn,
		fn: fn,
	}, nil

}

func (c *UDSPacketClient) SendMessage(data []byte) error {
	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to uds-seqpacket socket: %w", err)
	}
	return nil
}

func (c *UDSPacketClient) ReceiveMessage() ([]byte, error) {
	var buf [1024]byte
	n, err := c.conn.Read(buf[:])
	if err != nil {
		return nil, fmt.Errorf("error reading from uds-seqpacket socket: %w", err)
	}
	return buf[:n], nil
}

func (c *UDSPacketClient) Shutdown() error {
	err := c.conn.Close()
	os.Remove(c.fn)
	if err != nil {
		return fmt.Errorf("error closing connection on uds-seqpacket socket: %w", err)
	}
	return nil
}

func makeSocketFile(dstfn string) string {
	rand.Seed(time.Now().UnixNano())
	path, fn := filepath.Split(dstfn)
	ext := filepath.Ext(fn)
	withoutExt := strings.Replace(fn, ext, "", 1)
	newFn := fmt.Sprintf("%s_%d%s", withoutExt, rand.Int(), ext)
	return filepath.Join(path, newFn)
}

