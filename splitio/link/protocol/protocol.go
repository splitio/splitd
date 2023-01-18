package protocol

type Version byte

const (
	V1 Version = 0x01
)

type RPCBase struct {
	Version Version
}
