package transfer

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/splitio/go-toolkit/v5/logging"
)

type OnClientAttachedCallback = func(conn RawConn)
type RawConnFactory = func(conn net.Conn) RawConn

type Acceptor struct {
	onClientAttachedCallback OnClientAttachedCallback
	rawConnFactory RawConnFactory
	logger logging.LoggerInterface
	address net.Addr
}

func NewAcceptor(
	address net.Addr,
	onClientAttachedCallback OnClientAttachedCallback,
	rawConnFactory RawConnFactory,
	logger logging.LoggerInterface,
) *Acceptor {
	return &Acceptor{
		onClientAttachedCallback: onClientAttachedCallback,
		rawConnFactory: rawConnFactory,
		logger: logger,
		address: address,
	}
}

func (a *Acceptor) Start() (<-chan error, error) {
	l, err := net.Listen(a.address.Network(), a.address.String())
	if err != nil {
		return nil, fmt.Errorf("error listening on provided address: %w", err)
	}

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
			go a.onClientAttachedCallback(wrappedConn)
		}
	}()
	return ret, nil
}


