package sdk

import (
	"fmt"
	"testing"
	"time"

	"github.com/splitio/go-client/v6/splitio/engine/evaluator"
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/provisional"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage/inmemory"
	"github.com/splitio/go-split-commons/v4/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/splitio/splitd/splitio/sdk/workers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTreatmentLabelsDisabled(t *testing.T) {
	is, _ := storage.NewImpressionsQueue(100)

	ev := &EvaluatorMock{}
	ev.On("EvaluateFeature", "key1", (*string)(nil), "f1", Attributes{"a": 1}).
		Return(&evaluator.Result{Treatment: "on", Label: "label1", EvaluationTime: 1 * time.Millisecond, SplitChangeNumber: 123}).
		Once()

	expectedImpression := &dtos.Impression{
		KeyName:      "key1",
		BucketingKey: "",
		FeatureName:  "f1",
		Treatment:    "on",
		ChangeNumber: 123,
	}
	im := &ImpressionManagerMock{}
	im.On("ProcessSingle", mock.Anything).
		// hay que hacer el assert aca en lugar del matcher por el timestamp
		Run(func(a mock.Arguments) { assertImpEq(t, expectedImpression, a.Get(0).(*dtos.Impression)) }).
		Return(true).
		Once()

	client := &Impl{
		logger: logging.NewLogger(nil),
		is:     is,
		ev:     ev,
		iq:     im,
		cfg:    conf.Config{LabelsEnabled: false},
	}

	res, err := client.Treatment(&types.ClientConfig{Metadata: types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}}, "key1", nil, "f1", Attributes{"a": 1})
	assert.Nil(t, err)
	assert.Nil(t, res.Config)
	assertImpEq(t, expectedImpression, res.Impression)

	err = is.RangeAndClear(func(md types.ClientMetadata, st *storage.LockingQueue[dtos.Impression]) {
		assert.Equal(t, types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}, md)
		assert.Equal(t, 1, st.Len())

		var imps []dtos.Impression
		n, err := st.Pop(1, &imps)
		assert.Nil(t, nil)
		assert.Equal(t, 1, n)
		assert.Equal(t, 1, len(imps))
		assertImpEq(t, expectedImpression, &imps[0])
		n, err = st.Pop(1, &imps)
		assert.ErrorIs(t, err, storage.ErrQueueEmpty)

	})
	assert.Nil(t, err)
}

func TestTreatmentLabelsEnabled(t *testing.T) {
	is, _ := storage.NewImpressionsQueue(100)

	ev := &EvaluatorMock{}
	ev.On("EvaluateFeature", "key1", (*string)(nil), "f1", Attributes{"a": 1}).
		Return(&evaluator.Result{Treatment: "on", Label: "label1", EvaluationTime: 1 * time.Millisecond, SplitChangeNumber: 123}).
		Once()

	expectedImpression := &dtos.Impression{
		KeyName:      "key1",
		BucketingKey: "",
		FeatureName:  "f1",
		Treatment:    "on",
		Label:        "label1",
		ChangeNumber: 123,
	}
	im := &ImpressionManagerMock{}
	im.On("ProcessSingle", mock.Anything).
		Run(func(args mock.Arguments) {
			// hay que hacer el assert aca en lugar del matcher por el timestamp
			assertImpEq(t, expectedImpression, args.Get(0).(*dtos.Impression))
		}).
		Return(true).
		Once()

	client := &Impl{
		logger: logging.NewLogger(nil),
		is:     is,
		ev:     ev,
		iq:     im,
		cfg:    conf.Config{LabelsEnabled: true},
	}

	res, err := client.Treatment(&types.ClientConfig{Metadata: types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}}, "key1", nil, "f1", Attributes{"a": 1})
	assert.Nil(t, err)
	assert.Nil(t, res.Config)
	assertImpEq(t, expectedImpression, res.Impression)

	err = is.RangeAndClear(func(md types.ClientMetadata, st *storage.LockingQueue[dtos.Impression]) {
		assert.Equal(t, types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}, md)
		assert.Equal(t, 1, st.Len())

		var imps []dtos.Impression
		n, err := st.Pop(1, &imps)
		assert.Nil(t, nil)
		assert.Equal(t, 1, n)
		assert.Equal(t, 1, len(imps))
		assertImpEq(t, expectedImpression, &imps[0])
		n, err = st.Pop(1, &imps)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, storage.ErrQueueEmpty)

	})
	assert.Nil(t, err)
}

func TestTreatments(t *testing.T) {
	is, _ := storage.NewImpressionsQueue(100)

	ev := &EvaluatorMock{}
	ev.On("EvaluateFeatures", "key1", (*string)(nil), []string{"f1", "f2", "f3"}, Attributes{"a": 1}).
		Return(evaluator.Results{Evaluations: map[string]evaluator.Result{
			"f1": {Treatment: "on", Label: "label1", EvaluationTime: 1 * time.Millisecond, SplitChangeNumber: 123},
			"f2": {Treatment: "on", Label: "label2", EvaluationTime: 2 * time.Millisecond, SplitChangeNumber: 124},
			"f3": {Treatment: "on", Label: "label3", EvaluationTime: 3 * time.Millisecond, SplitChangeNumber: 125},
		}}).
		Once()

	expectedImpressions := []dtos.Impression{
		{KeyName: "key1", BucketingKey: "", FeatureName: "f1", Treatment: "on", Label: "label1", ChangeNumber: 123},
		{KeyName: "key1", BucketingKey: "", FeatureName: "f2", Treatment: "on", Label: "label2", ChangeNumber: 124},
		{KeyName: "key1", BucketingKey: "", FeatureName: "f3", Treatment: "on", Label: "label3", ChangeNumber: 125},
	}
	im := &ImpressionManagerMock{}
	im.On("ProcessSingle", mock.Anything).
		Run(func(args mock.Arguments) {
			assertImpEq(t, &expectedImpressions[0], args.Get(0).(*dtos.Impression))
		}).
		Return(true).
		Once()
	im.On("ProcessSingle", mock.Anything).
		Run(func(args mock.Arguments) {
			assertImpEq(t, &expectedImpressions[1], args.Get(0).(*dtos.Impression))
		}).
		Return(true).
		Once()
	im.On("ProcessSingle", mock.Anything).
		Run(func(args mock.Arguments) {
			assertImpEq(t, &expectedImpressions[2], args.Get(0).(*dtos.Impression))
		}).
		Return(true).
		Once()

	client := &Impl{
		logger: logging.NewLogger(nil),
		is:     is,
		ev:     ev,
		iq:     im,
		cfg:    conf.Config{LabelsEnabled: true},
	}

	res, err := client.Treatments(
		&types.ClientConfig{Metadata: types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}},
		"key1", nil, []string{"f1", "f2", "f3"}, Attributes{"a": 1})
	assert.Nil(t, err)
	assert.Nil(t, res["f1"].Config)
	assert.Nil(t, res["f2"].Config)
	assert.Nil(t, res["f3"].Config)
	assertImpEq(t, &expectedImpressions[0], res["f1"].Impression)
	assertImpEq(t, &expectedImpressions[1], res["f2"].Impression)
	assertImpEq(t, &expectedImpressions[2], res["f3"].Impression)

	err = is.RangeAndClear(func(md types.ClientMetadata, st *storage.LockingQueue[dtos.Impression]) {
		assert.Equal(t, types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}, md)
		assert.Equal(t, 3, st.Len())

		var imps []dtos.Impression
		n, err := st.Pop(3, &imps)
		assert.Nil(t, nil)
		assert.Equal(t, 3, n)
		assert.Equal(t, 3, len(imps))
		assertImpEq(t, &expectedImpressions[0], &imps[0])
		assertImpEq(t, &expectedImpressions[1], &imps[1])
		assertImpEq(t, &expectedImpressions[2], &imps[2])
		n, err = st.Pop(1, &imps)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, storage.ErrQueueEmpty)

	})
	assert.Nil(t, err)

}

func TestImpressionsQueueFull(t *testing.T) {

	logger := logging.NewLogger(nil)

	impRecorder := &ImpressionRecorderMock{}
	impRecorder.On("Record", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)

	// setup an impressions queue of 4, and a task with a large period for evicting, and a synchronizer
	// with only enough hata to flush impressions when the queue-full signal arrives
	// @{
	queueFullChan := make(chan string, 2)
	is, _ := storage.NewImpressionsQueue(4)
	ts, _ := inmemory.NewTelemetryStorage()
	iw := workers.NewImpressionsWorker(logger, ts, impRecorder, is, &conf.Impressions{Mode: "optimized", SyncPeriod: 100 * time.Second})
	sworkers := synchronizer.Workers{ImpressionRecorder: iw}
	sy := synchronizer.NewSynchronizer(*conf.DefaultConfig().ToAdvancedConfig(), synchronizer.SplitTasks{}, sworkers, logger, queueFullChan, nil)
	sy.StartPeriodicDataRecording()
	// @}

	ev := &EvaluatorMock{}
	ev.On("EvaluateFeature", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&evaluator.Result{Treatment: "on", Label: "label1", EvaluationTime: 1 * time.Millisecond, SplitChangeNumber: 123}).
		Times(9)

	expectedImpression := &dtos.Impression{KeyName: "key1", BucketingKey: "", FeatureName: "f1", Treatment: "on", Label: "label1", ChangeNumber: 123}
	im := &ImpressionManagerMock{}
	im.On("ProcessSingle", mock.Anything).
		// hay que hacer el assert aca en lugar del matcher por el timestamp
		Run(func(args mock.Arguments) { assertImpEq(t, expectedImpression, args.Get(0).(*dtos.Impression)) }).
		Return(true).
		Times(9)

	client := &Impl{logger: logging.NewLogger(nil), ss: nil, is: is, ev: ev, iq: im, cfg: conf.Config{LabelsEnabled: true}, queueFullChan: queueFullChan}
	clientConf := &types.ClientConfig{Metadata: types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}}

	// create 4 impressions to fill the queue (last one will be dropped and trigger a flush)
	for idx := 0; idx < 4; idx++ {
		feature := fmt.Sprintf("f%d", idx)
		expectedImpression.FeatureName = feature
		res, err := client.Treatment(clientConf, "key1", nil, feature, Attributes{"a": 1})
		assert.Nil(t, err)
		assert.Nil(t, res.Config)
		assertImpEq(t, expectedImpression, res.Impression)
	}

	time.Sleep(time.Second) // wait 1 sec to allow for a context switch since the flush is async

	// same
	for idx := 4; idx < 8; idx++ {
		feature := fmt.Sprintf("f%d", idx)
		expectedImpression.FeatureName = feature
		res, err := client.Treatment(clientConf, "key1", nil, feature, Attributes{"a": 1})
		assert.Nil(t, err)
		assert.Nil(t, res.Config)
		assertImpEq(t, expectedImpression, res.Impression)
	}

	time.Sleep(time.Second) // wait 1 sec to allow for a context switch since the flush is async

	feature := "f8"
	expectedImpression.FeatureName = feature
	res, err := client.Treatment(clientConf, "key1", nil, feature, Attributes{"a": 1})
	assert.Nil(t, err)
	assert.Nil(t, res.Config)
	assertImpEq(t, expectedImpression, res.Impression)

	ev.AssertExpectations(t)
	im.AssertExpectations(t)
	impRecorder.AssertExpectations(t)
	var totalSize int
	is.Range(func(md types.ClientMetadata, q *storage.LockingQueue[dtos.Impression]) { totalSize += q.Len() })
	assert.Equal(t, 1, totalSize) // assert no more impressions in queue
}

func assertImpEq(t *testing.T, i1, i2 *dtos.Impression) {
	t.Helper()
	assert.Equal(t, i1.KeyName, i2.KeyName)
	assert.Equal(t, i1.BucketingKey, i2.BucketingKey)
	assert.Equal(t, i1.FeatureName, i2.FeatureName)
	assert.Equal(t, i1.Treatment, i2.Treatment)
	assert.Equal(t, i1.Label, i2.Label)
	assert.Equal(t, i1.ChangeNumber, i2.ChangeNumber)
}

// mocks

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

var _ evaluator.Interface = (*EvaluatorMock)(nil)
var _ provisional.ImpressionManager = (*ImpressionManagerMock)(nil)
var _ service.ImpressionsRecorder = (*ImpressionRecorderMock)(nil)
