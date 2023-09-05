package serializer

import "fmt"

type Mechanism int

func (m Mechanism) String() string {
	switch m {
	case MsgPack:
		return "msgpack"
	default:
		return "invalid-serialization"
	}
}

const (
	MsgPack Mechanism = 1
)

func Setup(mechanism Mechanism) (Interface, error) {
	switch mechanism {
	case MsgPack:
		return newMessagePack(), nil
	}
	return nil, fmt.Errorf("unknown serialization mechanism '%d'", mechanism)
}
