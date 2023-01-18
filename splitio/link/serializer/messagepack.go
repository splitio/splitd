package serializer

import (
	"fmt"

	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"

	"github.com/vmihailenco/msgpack/v5"
)

type Interface[T rpcConstraint, U responseConstraint] interface {
	Parse([]byte) (*T, error)
	Serialize(U) ([]byte, error)
}

type MessagePack[T rpcConstraint, U responseConstraint] struct {
}

// Parse implements Interface
func (*MessagePack[T, U]) Parse(raw []byte) (*T, error) {
	var rpc T
	err := msgpack.Unmarshal(raw, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error deserializing rpc: %w", err)
	}
	return &rpc, nil
}

// Serialize implements Interface
func (*MessagePack[T, U]) Serialize(response U) ([]byte, error) {
	serialized, err := msgpack.Marshal(response.Get())
	if err != nil {
		return nil, fmt.Errorf("error serializing RPC: %w", err)
	}
	return serialized, nil
}

func newMessagePack[T rpcConstraint, U responseConstraint]() *MessagePack[T, U] {
	return &MessagePack[T, U]{}
}

type rpcConstraint interface {
	protov1.RPC
}

type responseConstraint interface {
	protov1.Response
}


var _ Interface[protov1.RPC, protov1.Response] = (*MessagePack[protov1.RPC, protov1.Response])(nil)
