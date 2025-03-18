package transfer

import (
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/splitio/go-toolkit/v5/logging"

	"github.com/iceber/iouring-go"
	iouring_syscall "github.com/iceber/iouring-go/syscall"
)

var (
	ErrNoMoreConns = errors.New("no more connections")
)

type ConnState byte

const (
	StateReadInProgress  ConnState = 0x02
	StateWriteInProgress ConnState = 0x11
)

type CSMFactoryFunc func() ClientStateMachine
type ClientStateMachine interface {
	HandleIncomingData(in []byte, out *[]byte) (int, error)
}

type IOURingLoop struct {
	logger     logging.LoggerInterface
	lAddr      string
	lfd        int
	csmFactory CSMFactoryFunc
	w          *iouring.IOURing
	conns      map[int]*connWrapper
	maxConns   int
}

type connWrapper struct { // fd -> conn
	csm      ClientStateMachine
	state    ConnState
	incoming []byte
	outgoing []byte
	respSize int
}

func (r *IOURingLoop) loop() {
	sqes := make([]iouring.PrepRequest, 0, 1024)
	results := make(chan iouring.Result, 1024)
	r.w.SubmitRequest(iouring.Accept(r.lfd), results)

	for {
		sqes = sqes[:0]
		handledCount := r.handlePendingEvents(results, &sqes)
		if handledCount == 0 {
			time.Sleep(1 * time.Millisecond) // TODO(mredolatti): how long to wait here?
		}
		// handle pending messages
		if len(sqes) > 0 {
			_, err := r.w.SubmitRequests(sqes, results)
			if err != nil {
				panic(err.Error()) // TODO
			}
		}
	}
}

func (r *IOURingLoop) handlePendingEvents(results <-chan iouring.Result, sqes *[]iouring.PrepRequest) int {
	operationsCompleted := 0
	for {
		select {
		case result, ok := <-results:
			if !ok {
				panic("CHAN CLOSED")
			}

			operationsCompleted++

			if err := result.Err(); err != nil {
				r.logger.Error(fmt.Sprintf("%+v\n", result))
			}

			switch result.Opcode() {
			case iouring_syscall.IORING_OP_ACCEPT:
				if err := r.handleAcceptDone(result, sqes); err != nil {
				}
			case iouring_syscall.IORING_OP_READ:
				if err := r.handleReadDone(result, sqes); err != nil {
				}
			case iouring_syscall.IORING_OP_WRITE:
				if err := r.handleWriteDone(result, sqes); err != nil {
				}
			case iouring_syscall.IORING_OP_CLOSE:
				// TODO!
				panic("conn closed!")

			default:
				panic("unexpected event")
			}
		default:
			return operationsCompleted
		}
	}
}

func (r *IOURingLoop) handleAcceptDone(result iouring.Result, sqes *[]iouring.PrepRequest) error {
	fd := result.ReturnValue0().(int)
	if l := len(r.conns); l >= r.maxConns {
		r.logger.Error("connection limit exceeded")
		return nil
	}

	cw := &connWrapper{
		csm:      r.csmFactory(),
		state:    StateReadInProgress,
		incoming: make([]byte, 1024),
		outgoing: make([]byte, 1024),
	}

	r.conns[fd] = cw
	r.logger.Info("started tracking fd=", fd)
	*sqes = append(*sqes, iouring.Accept(r.lfd), iouring.Read(fd, cw.incoming))
	return nil
}

func (r *IOURingLoop) handleReadDone(result iouring.Result, sqes *[]iouring.PrepRequest) error {
	forFd, ok := r.conns[result.Fd()]
	if !ok {
		panic(fmt.Sprintf("requested metadata for fd=%d which was not found", result.Fd()))
	}

	// TODO(mredolatti): this works for seqpacket-type sockets. stream-based ones will need more work.
	assertCurrentStatus(forFd, StateReadInProgress)
	bytesRead := result.ReturnValue0().(int)
	if bytesRead == 0 { // EOF
		delete(r.conns, result.Fd())
		return nil
	}

	n, err := forFd.csm.HandleIncomingData(forFd.incoming[:bytesRead], &forFd.outgoing)
	if err != nil {
		panic(fmt.Sprintf("error handling received data: %w", err))
	}

	*sqes = append(*sqes, iouring.Write(result.Fd(), forFd.outgoing[:n]))
	forFd.state = StateWriteInProgress // indicate that response is ready to be written now
	return nil
}

func (r *IOURingLoop) handleWriteDone(result iouring.Result, sqes *[]iouring.PrepRequest) error {
	forFd, ok := r.conns[result.Fd()]
	if !ok {
		panic(fmt.Sprintf("requested metadata for fd=%d which was not found", result.Fd()))
	}

	assertCurrentStatus(forFd, StateWriteInProgress)
	*sqes = append(*sqes, iouring.Read(result.Fd(), forFd.incoming))
	forFd.state = StateReadInProgress
	return nil

}

func NewIOUringLoop(
	addr string,
	csmFactory CSMFactoryFunc,
	logger logging.LoggerInterface,
	entries uint,
	maxConns int,
) (*IOURingLoop, error) {
	iou, err := iouring.New(entries)
	if err != nil {
		return nil, fmt.Errorf("error setting up io-uring: %w", err)
	}

	l := &IOURingLoop{
		csmFactory: csmFactory,
		lAddr:      addr,
		lfd:        listenSocket(addr),
		logger:     logger,
		w:          iou,
		conns:      make(map[int]*connWrapper, maxConns),
		maxConns:   maxConns,
	}

	go l.loop()
	return l, nil
}

func listenSocket(addr string) int {
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_SEQPACKET, 0)
	if err != nil {
		panic(err)
	}

	sockaddr := &syscall.SockaddrUnix{Name: addr}
	if err := syscall.Bind(fd, sockaddr); err != nil {
		panic(err)
	}

	if err := syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		panic(err)
	}
	return fd
}

func assertCurrentStatus(w *connWrapper, expected ConnState) {
	if w.state != expected {
		panic(fmt.Sprintf("expected status to be [%s] but was [%s]", expected, w.state))
	}
}
