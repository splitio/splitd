package parser

import (
	"fmt"

	"github.com/splitio/splitd/splitio/protocol"
	"github.com/vmihailenco/msgpack/v5"
)

type Interface interface {
	Parse([]byte) (*protocol.RPC, error)
}

type Impl struct {
}

// Parse implements Interface
func (*Impl) Parse(raw []byte) (*protocol.RPC, error) {
	var rpc protocol.RPC
	err := msgpack.Unmarshal(raw, &rpc)
	if err != nil {
		return nil, fmt.Errorf("error deserializing rpc: %w", err)
	}
	return &rpc, nil
}

func New() *Impl {
	return &Impl{}
}

var _ Interface = (*Impl)(nil)
