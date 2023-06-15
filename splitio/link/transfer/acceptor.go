package transfer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"golang.org/x/sync/semaphore"

	"github.com/splitio/go-toolkit/v5/logging"
)

const (
    defaultMaxSimultaneousConns = 32
)

type OnClientAttachedCallback = func(conn RawConn)
type RawConnFactory = func(conn net.Conn) RawConn

type Acceptor struct {
	listener atomic.Value
	rawConnFactory RawConnFactory
	logger logging.LoggerInterface
	address net.Addr
    sem *semaphore.Weighted
}

func newAcceptor(address net.Addr, rawConnFactory RawConnFactory, logger logging.LoggerInterface, maxConns int) *Acceptor {

    if maxConns == 0 {
        maxConns = defaultMaxSimultaneousConns
    }

	return &Acceptor{
		rawConnFactory: rawConnFactory,
		logger: logger,
		address: address,
        sem: semaphore.NewWeighted(int64(maxConns)),
	}
}

func (a *Acceptor) Start(onClientAttachedCallback OnClientAttachedCallback) (<-chan error, error) {
	l, err := net.Listen(a.address.Network(), a.address.String())
	if err != nil {
		return nil, fmt.Errorf("error listening on provided address: %w", err)
	}
	a.listener.Store(l)

	ret := make(chan error, 1)
	go func() {
		defer l.Close()
		for {

            a.sem.Acquire(context.Background(), 1)
			conn, err := l.Accept()
			if err != nil {
				var toSend error
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					toSend = err
				}
				ret <- toSend

				return
			}
			wrappedConn := a.rawConnFactory(conn)
			go func() {
                onClientAttachedCallback(wrappedConn)
                a.sem.Release(1)
            }()
		}
	}()
	return ret, nil
}

func (a *Acceptor) Shutdown() error {
	listener, ok := a.listener.Load().(net.Listener)
	if !ok {
		return nil // No listener set yet
	}

	err := listener.Close()
	if err != nil {
		return fmt.Errorf("error shutting down listener: %w", err)
	}

	return nil
}
