package sdk

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/splitio/go-client/v6/splitio/engine/evaluator"
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/storage/inmemory"
	"github.com/splitio/go-split-commons/v4/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/external/commons/mocks"
	"github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/splitio/splitd/splitio/sdk/workers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTreatmentLabelsDisabled(t *testing.T) {
	is, _ := storage.NewImpressionsQueue(100)

	ev := &mocks.EvaluatorMock{}
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
	im := &mocks.ImpressionManagerMock{}
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

	ev := &mocks.EvaluatorMock{}
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
	im := &mocks.ImpressionManagerMock{}
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

	ev := &mocks.EvaluatorMock{}
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
	im := &mocks.ImpressionManagerMock{}
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

	impRecorder := &mocks.ImpressionRecorderMock{}
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

	ev := &mocks.EvaluatorMock{}
	ev.On("EvaluateFeature", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&evaluator.Result{Treatment: "on", Label: "label1", EvaluationTime: 1 * time.Millisecond, SplitChangeNumber: 123}).
		Times(9)

	expectedImpression := &dtos.Impression{KeyName: "key1", BucketingKey: "", FeatureName: "f1", Treatment: "on", Label: "label1", ChangeNumber: 123}
	im := &mocks.ImpressionManagerMock{}
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

func TestTrack(t *testing.T) {

	es, _ := storage.NewEventsQueue(1000)
	logger := logging.NewLogger(nil)

	ss := &mocks.SplitStorageMock{}
	ss.On("TrafficTypeExists", "user").Return(true)

	client := &Impl{
		logger:    logging.NewLogger(nil),
		es:        es,
		cfg:       conf.Config{LabelsEnabled: false},
		validator: Validator{logger, ss},
	}

	md := types.ClientConfig{Metadata: types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}}

	err := client.Track(&md, "key1", "user", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.Nil(t, err)

	err = es.RangeAndClear(func(md types.ClientMetadata, st *storage.LockingQueue[dtos.EventDTO]) {
		assert.Equal(t, types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}, md)
		assert.Equal(t, 1, st.Len())

		var evs []dtos.EventDTO
		n, err := st.Pop(1, &evs)
		assert.Nil(t, nil)
		assert.Equal(t, 1, n)
		assert.Equal(t, 1, len(evs))
		assertEventEq(t, &dtos.EventDTO{
			Key:             "key1",
			TrafficTypeName: "user",
			EventTypeID:     "checkin",
			Value:           ref(123.4),
			Properties:      map[string]interface{}{"a": 123},
		}, &evs[0])
		n, err = st.Pop(1, &evs)
		assert.ErrorIs(t, err, storage.ErrQueueEmpty)

	})
	assert.Nil(t, err)

	err = client.Track(&md, "key1", "", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.ErrorIs(t, err, ErrEmtpyTrafficType)

	err = client.Track(&md, "key1", "user", "checkin", ref(123.4), map[string]interface{}{"a": strings.Repeat("qwertyui", 100000)})
	assert.ErrorIs(t, err, ErrEventTooBig)

}

func TestTrackEventsFlush(t *testing.T) {

	es, _ := storage.NewEventsQueue(4)
	logger := logging.NewLogger(nil)

	ss := &mocks.SplitStorageMock{}
	ss.On("TrafficTypeExists", "user").Return(true)

	client := &Impl{
		logger:        logging.NewLogger(nil),
		queueFullChan: make(chan string, 2),
		es:            es,
		cfg:           conf.Config{LabelsEnabled: false},
		validator:     Validator{logger, ss},
	}

	md := types.ClientConfig{Metadata: types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}}

	err := client.Track(&md, "key1", "user", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.Nil(t, err)
	err = client.Track(&md, "key2", "user", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.Nil(t, err)
	err = client.Track(&md, "key3", "user", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.Nil(t, err)
	err = client.Track(&md, "key4", "user", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.ErrorIs(t, err, ErrEventsQueueFull)

	assert.Equal(t, "EVENTS_FULL", <-client.queueFullChan)

	err = es.RangeAndClear(func(md types.ClientMetadata, st *storage.LockingQueue[dtos.EventDTO]) {
		assert.Equal(t, types.ClientMetadata{ID: "some", SdkVersion: "go-1.2.3"}, md)
		assert.Equal(t, 3, st.Len())

		var evs []dtos.EventDTO
		n, err := st.Pop(10, &evs)
		assert.Nil(t, nil)
		assert.Equal(t, 3, n)
		assert.Equal(t, 3, len(evs))

		expectedEvent := dtos.EventDTO{
			Key:             "key1",
			TrafficTypeName: "user",
			EventTypeID:     "checkin",
			Value:           ref(123.4),
			Properties:      map[string]interface{}{"a": 123},
		}

		assertEventEq(t, &expectedEvent, &evs[0])
		expectedEvent.Key = "key2"
		assertEventEq(t, &expectedEvent, &evs[1])
		expectedEvent.Key = "key3"
		assertEventEq(t, &expectedEvent, &evs[2])

		n, err = st.Pop(1, &evs)
		assert.ErrorIs(t, err, storage.ErrQueueEmpty)

	})
	assert.Nil(t, err)

	err = client.Track(&md, "key1", "", "checkin", ref(123.4), map[string]interface{}{"a": 123})
	assert.ErrorIs(t, err, ErrEmtpyTrafficType)

	err = client.Track(&md, "key1", "user", "checkin", ref(123.4), map[string]interface{}{"a": strings.Repeat("qwertyui", 100000)})
	assert.ErrorIs(t, err, ErrEventTooBig)

}

func TestSplitNames(t *testing.T) {
	var ss mocks.SplitStorageMock
	ss.On("SplitNames").Return([]string{"split1", "split2"}).Once()

	c := Impl{splitStorage: &ss}

	names, err := c.SplitNames()
	assert.Nil(t, err)
	assert.Equal(t, []string{"split1", "split2"}, names)
}

func TestSplits(t *testing.T) {
	var ss mocks.SplitStorageMock
	ss.On("All").Return([]dtos.SplitDTO{
		{Name: "s1", TrafficTypeName: "tt1", ChangeNumber: 1, Conditions: []dtos.ConditionDTO{{Partitions: []dtos.PartitionDTO{{Treatment: "a"}, {Treatment: "b"}}}}},
		{
			Name:            "s2",
			TrafficTypeName: "tt1",
			ChangeNumber:    1,
			Conditions:      []dtos.ConditionDTO{{Partitions: []dtos.PartitionDTO{{Treatment: "a"}, {Treatment: "b"}}}},
			Configurations:  map[string]string{"a": "conf1", "b": "conf2"},
		},
	}).Once()

	c := Impl{splitStorage: &ss}

	splits, err := c.Splits()
	assert.Nil(t, err)
	assert.Equal(t, []SplitView{
		{Name: "s1", TrafficType: "tt1", Killed: false, Treatments: []string{"a", "b"}, ChangeNumber: 1},
		{Name: "s2", TrafficType: "tt1", Killed: false, Treatments: []string{"a", "b"}, ChangeNumber: 1, Configs: map[string]string{"a": "conf1", "b": "conf2"}},
	}, splits)
}

func TestSplit(t *testing.T) {
	var ss mocks.SplitStorageMock
	ss.On("Split", "s2").Return(&dtos.SplitDTO{
		Name:            "s2",
		TrafficTypeName: "tt1",
		ChangeNumber:    1,
		Conditions:      []dtos.ConditionDTO{{Partitions: []dtos.PartitionDTO{{Treatment: "a"}, {Treatment: "b"}}}},
		Configurations:  map[string]string{"a": "conf1", "b": "conf2"},
	}).Once()

	c := Impl{splitStorage: &ss}

	split, err := c.Split("s2")
	assert.Nil(t, err)
	assert.Equal(t, &SplitView{
		Name:         "s2",
		TrafficType:  "tt1",
		Killed:       false,
		Treatments:   []string{"a", "b"},
		ChangeNumber: 1,
		Configs:      map[string]string{"a": "conf1", "b": "conf2"},
	}, split)
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

func assertEventEq(t *testing.T, e1, e2 *dtos.EventDTO) {
	t.Helper()
	assert.Equal(t, e1.Key, e2.Key)
	assert.Equal(t, e1.TrafficTypeName, e2.TrafficTypeName)
	assert.Equal(t, e1.EventTypeID, e2.EventTypeID)
	assert.Equal(t, e1.Value, e2.Value)
	assert.Equal(t, e1.Properties, e2.Properties)
}

func ref[T any](t T) *T {
	return &t
}
