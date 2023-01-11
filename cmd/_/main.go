package main

import (
	"fmt"
	"net"
	"os"

	"github.com/splitio/split-agent/common"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {

	if len(os.Args) < 2 {
		panic("need at least one argument")
	}

	t := "unixpacket"
	laddr := net.UnixAddr{Name: "/tmp/unixdomaincli", Net: t}
	conn, err := net.DialUnix(t, &laddr, &net.UnixAddr{
		Name: "/tmp/unixdomain",
		Net:  t,
	})
	if err != nil {
		panic(err)
	}
	defer os.Remove("/tmp/unixdomaincli")

	serialized, err := msgpack.Marshal(common.Message{Func: "print", Args: []interface{}{os.Args[1], 3, true}})
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(serialized)
	if err != nil {
		panic(err)
	}

	var buf [1024]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		panic(err)
	}

	fmt.Println("response: " + string(buf[:n]))

	conn.Close()

}
