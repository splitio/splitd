package mocks

import (
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/provisional"
	"github.com/stretchr/testify/mock"
)

type ImpressionManagerMock struct {
	mock.Mock
}

// ProcessImpressions implements provisional.ImpressionManager
func (m *ImpressionManagerMock) ProcessImpressions(impressions []dtos.Impression) ([]dtos.Impression, []dtos.Impression) {
	args := m.Called(impressions)
	return args.Get(0).([]dtos.Impression), args.Get(1).([]dtos.Impression)
}

// ProcessSingle implements provisional.ImpressionManager
func (m *ImpressionManagerMock) ProcessSingle(impression *dtos.Impression) bool {
	args := m.Called(impression)
	return args.Bool(0)
}

func (m *ImpressionManagerMock) Process(impressions []dtos.Impression, listenerEnabled bool) ([]dtos.Impression, []dtos.Impression) {
	args := m.Called(impressions)
	return args.Get(0).([]dtos.Impression), args.Get(1).([]dtos.Impression)
}

var _ provisional.ImpressionManager = (*ImpressionManagerMock)(nil)
