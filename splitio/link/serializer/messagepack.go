package serializer

import (
	"github.com/vmihailenco/msgpack/v5"
)

type Interface interface {
	Parse([]byte, interface{}) error
	Serialize(interface{}) ([]byte, error)
}

type MessagePack struct {
}

// Parse implements Interface
func (*MessagePack) Parse(raw []byte, obj interface{}) error {
	return msgpack.Unmarshal(raw, &obj)
}

// Serialize implements Interface
func (*MessagePack) Serialize(obj interface{}) ([]byte, error) {
	return msgpack.Marshal(obj)
}

func newMessagePack() *MessagePack {
	return &MessagePack{}
}

var _ Interface = (*MessagePack)(nil)
