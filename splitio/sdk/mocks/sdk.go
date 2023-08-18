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
) (*sdk.EvaluationResult, error) {
	args := m.Called(md, key, bucketingKey, feature, attributes)
	return args.Get(0).(*sdk.EvaluationResult), args.Error(1)
}

// Treatments implements sdk.Interface
func (m *SDKMock) Treatments(
	md *types.ClientConfig,
	key string,
	bucketingKey *string,
	features []string,
	attributes map[string]interface{},
) (map[string]sdk.EvaluationResult, error) {
	args := m.Called(md, key, bucketingKey, features, attributes)
	return args.Get(0).(map[string]sdk.EvaluationResult), args.Error(1)
}

// Track implements sdk.Interface
func (m *SDKMock) Track(cfg *types.ClientConfig, key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error {
    args := m.Called(cfg, key, trafficType, eventType, value, properties)
    return args.Error(0)
}

var _ sdk.Interface = (*SDKMock)(nil)
