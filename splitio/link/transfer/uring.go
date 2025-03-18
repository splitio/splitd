package transfer

import (
	"errors"
	"fmt"
	"sync"
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
	StatePendingRead     ConnState = 0x01
	StateReadInProgress  ConnState = 0x02
	StatePendingWrite    ConnState = 0x10
	StateWriteInProgress ConnState = 0x11
)

type CSMFactoryFunc func() ClientStateMachine
type ClientStateMachine interface {
	HandleIncomingData(in []byte, out *[]byte) (int, error)
}

type IOURingLoop struct {
	logger     logging.LoggerInterface
	lAddr      string
	csmFactory CSMFactoryFunc
	w          *iouring.IOURing
	conns      map[int]*connWrapper
	mtx        sync.Mutex
	maxConns   int
}

type connWrapper struct { // fd -> conn
	csm      ClientStateMachine
	state    ConnState
	c        RawConn
	incoming []byte
	outgoing []byte
	respSize int
}

func (r *IOURingLoop) TrackConnection(c RawConn, csm ClientStateMachine) error {

	fd, err := c.FD()
	if err != nil {
		return fmt.Errorf("error obtaining file descriptor for new connection: %w", err)
	}
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if l := len(r.conns); l >= r.maxConns {
		return ErrNoMoreConns
	}

	r.conns[fd] = &connWrapper{
		c:        c,
		csm:      csm,
		state:    StatePendingRead,
		incoming: make([]byte, 1024),
		outgoing: make([]byte, 1024),
	}

	r.logger.Info("started tracking fd=", fd)
	return nil

}

func (r *IOURingLoop) loop() {
	sqes := make([]iouring.PrepRequest, 0, 1024)
	ch := make(chan iouring.Result, 1024)

	lfd := listenSocket(r.lAddr)
	r.w.SubmitRequest(iouring.Accept(lfd), ch)

	for {
		sqes = sqes[:0]
		r.mtx.Lock()
		for fd, wc := range r.conns {
			switch wc.state {
			case StatePendingRead:
				r.logger.Info("scheduling read for fd=", fd)
				sqes = append(sqes, iouring.Read(fd, wc.incoming))
				wc.state = StateReadInProgress

			case StatePendingWrite:
				r.logger.Info("scheduling write for fd=", fd)
				sqes = append(sqes, iouring.Write(fd, wc.outgoing[:wc.respSize]))
				wc.state = StateWriteInProgress
			}
		}
		r.mtx.Unlock()

		if len(sqes) > 0 {
			r.logger.Info("scheduling transfer requests")
			_, err := r.w.SubmitRequests(sqes, ch)
			if err != nil {
				panic(err.Error()) // TODO
			}
		}
		r.awaitAndHandleCompletions(ch)
	}

}

func (r *IOURingLoop) awaitAndHandleCompletions(ch chan iouring.Result) {
	operationsCompleted := 0
	for {
		select {
		case result, ok := <-ch:
			if !ok {
				panic("CHAN CLOSED")
			}
			operationsCompleted++

			if err := result.Err(); err != nil {
				panic(err.Error())
			}

			forFd, ok := r.conns[result.Fd()]
			if !ok {
				panic(fmt.Sprintf("requested metadata for fd=%d which was not found", result.Fd()))
			}

			switch result.Opcode() {
			case iouring_syscall.IORING_OP_ACCEPT:
				panic("ACCEPTED!")
			case iouring_syscall.IORING_OP_READ:
				r.logger.Info("READ COMPLETED")
				assertCurrentStatus(forFd, StateReadInProgress)
				bytesRead := result.ReturnValue0().(int)
				if bytesRead == 0 {
					// possibly connection closed. retry read
					forFd.state = StatePendingRead
					continue
				}

				n, err := forFd.csm.HandleIncomingData(forFd.incoming[:bytesRead], &forFd.outgoing)
				if err != nil {
					panic(fmt.Sprintf("error handling received data: %w", err))
				}

				forFd.state = StatePendingWrite // indicate that response is ready to be written now
				forFd.respSize = n

			case iouring_syscall.IORING_OP_WRITE:
				r.logger.Info("WRITE COMPLETED")
				assertCurrentStatus(forFd, StateWriteInProgress)
				forFd.state = StatePendingRead // ready for accepting next op
				forFd.respSize = 0

			case iouring_syscall.IORING_OP_CLOSE:
				if err := forFd.c.Shutdown(); err != nil {
					r.logger.Error("error shutting down client connection: %w", err)
				}

			default:
				panic("unexpected event")
			}
		default:
			if operationsCompleted == 0 {
				// nothing got completed so far. wait a while before re-iterating
				time.Sleep(100 * time.Millisecond)
			}
			return
		}
	}
}

func NewIOUringLoop(addr string, logger logging.LoggerInterface, entries uint, maxConns int) (*IOURingLoop, error) {
	iou, err := iouring.New(entries)
	if err != nil {
		return nil, fmt.Errorf("error setting up io-uring: %w", err)
	}

	l := &IOURingLoop{
		lAddr:    addr,
		logger:   logger,
		w:        iou,
		conns:    make(map[int]*connWrapper, maxConns),
		maxConns: maxConns,
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
