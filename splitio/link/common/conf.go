package common

import (
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
)

type Opts struct {
	ProtoV protocol.Version
	Serial serializer.Mechanism
}

func (o *Opts) Parse(os []Option) {
	for _, configure := range os {
		configure(o)
	}
}

func DefaultOpts() Opts {
	return Opts{
		ProtoV: protocol.V1,
		Serial: serializer.MsgPack,
	}
}

type Option func(*Opts)

func WithProtocolV(v protocol.Version) Option         { return func(o *Opts) { o.ProtoV = v } }
func WithSerialization(m serializer.Mechanism) Option { return func(o *Opts) { o.Serial = m } }
