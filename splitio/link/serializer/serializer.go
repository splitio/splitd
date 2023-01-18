package serializer

import "fmt"

type Mechanism int
const (
	MsgPack Mechanism = 1
)

func Setup[T rpcConstraint, U responseConstraint](mechanism Mechanism) (Interface[T, U], error) {
	switch mechanism {
	case MsgPack:
		return newMessagePack[T, U](), nil
	}
	return nil, fmt.Errorf("unknown serialization mechanism '%d'", mechanism)
}
