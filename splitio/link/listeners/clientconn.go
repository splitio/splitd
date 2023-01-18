package listeners

const (
	bufSize int = 1024
)

type ClientConnection interface {
	SendMessage([]byte) error
	ReceiveMessage() ([]byte, error)
	Shutdown() error
}
