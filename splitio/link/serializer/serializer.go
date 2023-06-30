package serializer

import "fmt"

type Mechanism int
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
