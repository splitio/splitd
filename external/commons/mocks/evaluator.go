package mocks

import (
	"github.com/splitio/go-split-commons/v5/engine/evaluator"
	"github.com/stretchr/testify/mock"
)

type EvaluatorMock struct {
	mock.Mock
}

// EvaluateFeature implements evaluator.Interface
func (e *EvaluatorMock) EvaluateFeature(key string, bucketingKey *string, feature string, attributes map[string]interface{}) *evaluator.Result {
	args := e.Called(key, bucketingKey, feature, attributes)
	return args.Get(0).(*evaluator.Result)
}

// EvaluateFeatures implements evaluator.Interface
func (e *EvaluatorMock) EvaluateFeatures(key string, bucketingKey *string, features []string, attributes map[string]interface{}) evaluator.Results {
	args := e.Called(key, bucketingKey, features, attributes)
	return args.Get(0).(evaluator.Results)
}

func (e *EvaluatorMock) EvaluateFeatureByFlagSets(key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) evaluator.Results {
	args := e.Called(key, bucketingKey, flagSets, attributes)
	return args.Get(0).(evaluator.Results)
}

var _ evaluator.Interface = (*EvaluatorMock)(nil)
