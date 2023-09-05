package mocks

import (
	"github.com/splitio/splitd/splitio/link/transfer"
	"github.com/stretchr/testify/mock"
)

type RawConnMock struct {
	mock.Mock
}

// ReceiveMessage implements transfer.RawConn
func (m *RawConnMock) ReceiveMessage() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// SendMessage implements transfer.RawConn
func (m *RawConnMock) SendMessage(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

// Shutdown implements transfer.RawConn
func (m *RawConnMock) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

var _ transfer.RawConn = (*RawConnMock)(nil)
