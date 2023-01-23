package transfer

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/splitio/go-toolkit/v5/logging"
)

type OnClientAttachedCallback = func(conn RawConn)
type RawConnFactory = func(conn net.Conn) RawConn

type Acceptor struct {
	listener atomic.Value
	rawConnFactory RawConnFactory
	logger logging.LoggerInterface
	address net.Addr
}

func newAcceptor(address net.Addr, rawConnFactory RawConnFactory, logger logging.LoggerInterface) *Acceptor {
	return &Acceptor{
		rawConnFactory: rawConnFactory,
		logger: logger,
		address: address,
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
			go onClientAttachedCallback(wrappedConn)
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
