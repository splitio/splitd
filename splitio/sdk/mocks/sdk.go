package mocks

import (
	"github.com/splitio/splitd/splitio/sdk"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/stretchr/testify/mock"
)

type SDKMock struct {
	mock.Mock
}

// Treatment implements sdk.Interface
func (m *SDKMock) Treatment(
	md *types.ClientConfig,
	key string,
	bucketingKey *string,
	feature string,
	attributes map[string]interface{},
) (*sdk.Result, error) {
	args := m.Called(md, key, bucketingKey, feature, attributes)
	return args.Get(0).(*sdk.Result), nil
}

// Treatments implements sdk.Interface
func (m *SDKMock) Treatments(
	md *types.ClientConfig,
	key string,
	bucketingKey *string,
	features []string,
	attributes map[string]interface{},
) (map[string]sdk.Result, error) {
	args := m.Called(md, key, bucketingKey, features, attributes)
	return args.Get(0).(map[string]sdk.Result), nil
}


var _ sdk.Interface = (*SDKMock)(nil)
