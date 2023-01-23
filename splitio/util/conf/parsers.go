package conf

import (
	"fmt"

	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
)

func ParseProtocolVersion(p string) (protocol.Version, error) {
	switch p {
	case "v1":
		return protocol.V1, nil
	}
	return 0, fmt.Errorf("unkown protocol version '%s'", p)
}

func ParseConnType(t string) (transfer.ConnType, error) {
	switch t {
	case "unix-seqpacket":
		return transfer.ConnTypeUnixSeqPacket, nil
	case "unix-stream":
		return transfer.ConnTypeUnixStream, nil
	}
	return 0, fmt.Errorf("unknown listener type '%s'", t)
}

func ParseSerializer(s string) (serializer.Mechanism, error) {
	switch s {
	case "msgpack":
		return serializer.MsgPack, nil
	}
	return 0, fmt.Errorf("unknown serialization mechanism '%s'", s)
}
