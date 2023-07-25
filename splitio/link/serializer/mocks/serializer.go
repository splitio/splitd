package mocks

import (
	"github.com/splitio/splitd/splitio/link/serializer"
	"github.com/stretchr/testify/mock"
)

type SerializerMock struct {
	mock.Mock
}

// Parse implements serializer.Interface
func (m *SerializerMock) Parse(data []byte, v interface{}) error {
	args := m.Called(data, v)
	return args.Error(0)
}

// Serialize implements serializer.Interface
func (m *SerializerMock) Serialize(v interface{}) ([]byte, error) {
	args := m.Called(v)
	return args.Get(0).([]byte), args.Error(1)
}

var _ serializer.Interface = (*SerializerMock)(nil)
