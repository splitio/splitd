package consumer

import "errors"

type ClientConnType int
const (
	ConnTypeUnixSeqPacket ClientConnType = 1
)

var (
	ErrInvalidClientConnType = errors.New("invalid Client type")
)

type Option func(*options)
func WithFileName(fn string) Option { return func(o *options) { o.fn = fn } }

type Raw interface {
	SendMessage(message []byte) error
	ReceiveMessage() ([]byte, error)
	Shutdown() error
}

func Create(lt ClientConnType, opts ...Option) (Raw, error) {
	
	var options options
	for _, apply := range opts {
		apply(&options)
	}

	switch lt {
	case ConnTypeUnixSeqPacket:
		return newUDSPacketClient(&options)
	default:
		return nil, ErrInvalidClientConnType
	}
}

type options struct {
	fn string
}
