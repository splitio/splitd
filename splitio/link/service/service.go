package service

import (
	"errors"
	"fmt"

	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/link/common"
	"github.com/splitio/splitd/splitio/link/protocol"
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/splitio/splitd/splitio/sdk"

	serviceV1 "github.com/splitio/splitd/splitio/link/service/v1"
)

var (
	ErrParsingData = errors.New("error parsing incoming message")
)

type Interface interface {
	HandleNewClient(cc transfer.RawConn)
}

type Impl struct {
	logger           logging.LoggerInterface
	splitSDK         sdk.Interface
	newClientManager ClientManagerFactory
}

func (s *Impl) HandleNewClient(cc transfer.RawConn) {
	cm := s.newClientManager(cc)
	cm.Manage()
	// TODO(mredolatti): Track active connections
}

func New(logger logging.LoggerInterface, splitSDK sdk.Interface, opts ...common.Option) (*Impl, error) {
	o := common.DefaultOpts()
	o.Parse(opts)

	switch o.ProtoV {
	case protocol.V1:
		cmf, err := newCMFactoryForV1(logger, splitSDK, o.Serial)
		if err != nil {
			return nil, fmt.Errorf("error setting up client-manager factory: %w", err)
		}
		return &Impl{
			logger:           logger,
			splitSDK:         splitSDK,
			newClientManager: cmf,
		}, nil
	}
	return nil, fmt.Errorf("unknown protocol version: '%d'", o.ProtoV)
}

type ClientManager interface {
	Manage()
}

type ClientManagerFactory func(transfer.RawConn) ClientManager

func newCMFactoryForV1(logger logging.LoggerInterface, splitSDK sdk.Interface, serialization serializer.Mechanism) (ClientManagerFactory, error) {
	ser, err := serializer.Setup(serialization)
	if err != nil {
		return nil, err
	}
	return func(conn transfer.RawConn) ClientManager {
		return serviceV1.NewClientManager(conn, logger, splitSDK, ser)
	}, nil
}

var _ Interface = (*Impl)(nil)
var _ ClientManager = (*serviceV1.ClientManager)(nil)
