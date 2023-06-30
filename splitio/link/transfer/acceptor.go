package transfer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	defaultMaxSimultaneousConns = 32
)

type OnClientAttachedCallback = func(conn RawConn)
type RawConnFactory = func(conn net.Conn) RawConn

type Acceptor struct {
	listener       atomic.Value
	rawConnFactory RawConnFactory
	logger         logging.LoggerInterface
	address        net.Addr
	maxConns       int
	sem            *semaphore.Weighted
	maxWait        time.Duration
}

var errNoSetDeadline = errors.New("listener doesn't support setting a deadline")

func newAcceptor(address net.Addr, rawConnFactory RawConnFactory, o *Options) *Acceptor {
	return &Acceptor{
		rawConnFactory: rawConnFactory,
		logger:         o.Logger,
		address:        address,
		maxConns:       o.MaxSimultaneousConnections,
		sem:            semaphore.NewWeighted(int64(o.MaxSimultaneousConnections)),
		maxWait:        o.AcceptTimeout,
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
			err = setDeadline(l, time.Now().Add(a.maxWait))
			if err != nil {
				a.logger.Warning("failed to set deadline for Accept call: ", err.Error())
			}

			conn, err := l.Accept()
			if err != nil {
				if os.IsTimeout(err) {
					// This just means that no-body tried to connect. ignore
					// The timeout is needed so that the loop can eventually break if
					// we trigger a shutdown and no-one is actually trying to connect
					continue
				}

				var toSend error
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					// This happens when accept fails for reasons other than a manual
					// shutdown request
					toSend = err
				}
				ret <- toSend
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), a.maxWait)
			err = a.sem.Acquire(ctx, 1)
			if err != nil {
				a.logger.Error(fmt.Sprintf("Incoming connection request timed out. If the current parallelism is expected, "+
					"consider increasing `maxConcurrentConnections` (current=%d)", a.maxConns))
				conn.Close()
				continue
			}
			cancel() // connection allowed. abort timeout timer

			go func(rc RawConn) {
				onClientAttachedCallback(rc)
				rc.Shutdown()
				a.sem.Release(1)
			}(a.rawConnFactory(conn))
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

// -- small helper & interface to uniformly set deadlines on different types of sockets
// @{

func setDeadline(listener net.Listener, deadline time.Time) error {
	if l, ok := listener.(hasSetDeadLine); ok {
		return l.SetDeadline(deadline)
	}
	return errNoSetDeadline
}

type hasSetDeadLine interface{ SetDeadline(t time.Time) error }

// @}
