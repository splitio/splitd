package protocol

type Version byte

func (v Version) String() string {
	switch v {
	case V1:
		return "v1"
	default:
		return "invalid-version"
	}
}

const (
	V1 Version = 0x01
)

type RPCBase struct {
	Version Version `msgpack:"v"`
}
