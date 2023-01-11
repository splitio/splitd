package listeners

import "errors"

type ListenerType int

const (
	ListenerTypeUnixSeqPacket = 1
)

type OnRawMessageCallback = func([]byte) ([]byte, error)

var (
	ErrInvalidListenerType = errors.New("invalid listener type")
)

type Option func(*options)
func WithFileName(fn string) Option { return func(o *options) { o.fn = fn } }

type Raw interface {
	Listen(onMessage OnRawMessageCallback) <-chan error
}

func Create(lt ListenerType, opts ...Option) (Raw, error) {
	
	var options options
	for _, apply := range opts {
		apply(&options)
	}

	switch lt {
	case ListenerTypeUnixSeqPacket:
		return newUDSSeqPacketListener(&options)
	default:
		return nil, ErrInvalidListenerType
	}
}

type options struct {
	fn string
}
