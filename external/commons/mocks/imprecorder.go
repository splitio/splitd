package mocks

import (
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/service"
	"github.com/stretchr/testify/mock"
)

type ImpressionRecorderMock struct {
	mock.Mock
}

// Record implements service.ImpressionsRecorder
func (m *ImpressionRecorderMock) Record(impressions []dtos.ImpressionsDTO, metadata dtos.Metadata, extraHeaders map[string]string) error {
	args := m.Called(impressions, metadata, extraHeaders)
	return args.Error(0)
}

// RecordImpressionsCount implements service.ImpressionsRecorder
func (m *ImpressionRecorderMock) RecordImpressionsCount(pf dtos.ImpressionsCountDTO, metadata dtos.Metadata) error {
	args := m.Called(pf, metadata)
	return args.Error(0)
}

var _ service.ImpressionsRecorder = (*ImpressionRecorderMock)(nil)
