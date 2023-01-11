package main

import (
	"fmt"

	"github.com/splitio/splitd/splitio/listeners"
)

func main() {
	l, err := listeners.Create(
		listeners.ListenerTypeUnixSeqPacket,
		listeners.WithFileName("./test.sock"),
	)

	if err != nil {
		panic(err.Error())
	}

	ec := l.Listen(func(recv []byte) ([]byte, error) {
		fmt.Println(string(recv))
		return []byte("hola"), nil
	})

	err = <-ec
	fmt.Println(err.Error())
}
