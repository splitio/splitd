package service

import (
	"errors"
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/listeners"
	"github.com/splitio/splitd/splitio/link/protocol"
	protov1 "github.com/splitio/splitd/splitio/link/protocol/v1"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/sdk"

	serviceV1 "github.com/splitio/splitd/splitio/link/service/v1"
)

var (
	ErrParsingData = errors.New("error parsing incoming message")
)

type Interface interface {
	HandleNewClient(cc listeners.ClientConnection)
}

type Impl struct {
	logger           logging.LoggerInterface
	splitSDK         sdk.Interface
	newClientManager ClientManagerFactory
}

func (s *Impl) HandleNewClient(cc listeners.ClientConnection) {
	cm := s.newClientManager(cc)
	go cm.Manage()
	// TODO(mredolatti): Track active connections
}

func New(logger logging.LoggerInterface, splitSDK sdk.Interface, version protocol.Version, sm serializer.Mechanism) (*Impl, error) {
	switch version {
	case protocol.V1:
		cmf, err := newCMFactoryForV1(logger, splitSDK, sm)
		if err != nil {
			return nil, fmt.Errorf("error setting up client-manager factory: %w", err)
		}
		return &Impl{
			logger:           logger,
			splitSDK:         splitSDK,
			newClientManager: cmf,
		}, nil
	}
	return nil, fmt.Errorf("unknown protocol version: '%d'", version)
}

type ClientManager interface {
	Manage()
}

type ClientManagerFactory func(listeners.ClientConnection) ClientManager

func newCMFactoryForV1(logger logging.LoggerInterface, splitSDK sdk.Interface, serialization serializer.Mechanism) (ClientManagerFactory, error) {
	ser, err := serializer.Setup[protov1.RPC, protov1.Response](serialization)
	if err != nil {
		return nil, err
	}
	return func(conn listeners.ClientConnection) ClientManager {
		return serviceV1.NewClientManager(conn, logger, splitSDK, ser)
	}, nil
}

var _ Interface = (*Impl)(nil)
var _ ClientManager = (*serviceV1.ClientManager)(nil)
