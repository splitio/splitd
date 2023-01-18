package listeners

import (
	"errors"

	"github.com/splitio/go-toolkit/v5/logging"
)

type ListenerType int

const (
	ListenerTypeUnixSeqPacket ListenerType = 1
	ListenerTypeUnixStream    ListenerType = 2
)

type OnClientAttachedCallback = func(conn ClientConnection)
type OnRawMessageCallback = func([]byte) ([]byte, error)

var (
	ErrInvalidListenerType = errors.New("invalid listener type")
	ErrSentDataMismatch    = errors.New("sent data size mismatch")
	ErrBufferTooSmall      = errors.New("insufficient capacity in read buffer")
)

type Option func(*options)

func WithFileName(fn string) Option                    { return func(o *options) { o.fn = fn } }
func WithLogger(logger logging.LoggerInterface) Option { return func(o *options) { o.logger = logger } }

type Raw interface {
	Listen(onMessage OnClientAttachedCallback) <-chan error
	Shutdown() error
}

func Create(lt ListenerType, opts ...Option) (Raw, error) {

	var o options
	for _, apply := range opts {
		apply(&o)
	}

	if o.logger == nil {
		o.logger = logging.NewLogger(nil)
	}

	switch lt {
	case ListenerTypeUnixSeqPacket:
		return newUDSSeqPacketListener(&o)
	case ListenerTypeUnixStream:
		return newUDSStreamListener(&o)
	default:
		return nil, ErrInvalidListenerType
	}
}

type options struct {
	fn     string
	logger logging.LoggerInterface
}
