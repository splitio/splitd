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

// TreatmentsByFlagSet implements sdk.Interface
func (m *SDKMock) TreatmentsByFlagSet(
	md *types.ClientConfig,
	key string,
	bucketingKey *string,
	flagSet string,
	attributes map[string]interface{},
) (map[string]sdk.EvaluationResult, error) {
	args := m.Called(md, key, bucketingKey, flagSet, attributes)
	return args.Get(0).(map[string]sdk.EvaluationResult), args.Error(1)
}

// Track implements sdk.Interface
func (m *SDKMock) Track(cfg *types.ClientConfig, key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error {
	args := m.Called(cfg, key, trafficType, eventType, value, properties)
	return args.Error(0)
}

func (m *SDKMock) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

// Split implements sdk.Interface
func (m *SDKMock) Split(name string) (*sdk.SplitView, error) {
	args := m.Called(name)
	return args.Get(0).(*sdk.SplitView), args.Error(1)
}

// SplitNames implements sdk.Interface
func (m *SDKMock) SplitNames() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// Splits implements sdk.Interface
func (m *SDKMock) Splits() ([]sdk.SplitView, error) {
	args := m.Called()
	return args.Get(0).([]sdk.SplitView), args.Error(1)
}

var _ sdk.Interface = (*SDKMock)(nil)
