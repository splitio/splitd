package workers

import (
	"testing"
	"time"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage/inmemory"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventsTask(t *testing.T) {
	is, _ := sss.NewEventsQueue(100)
	ts, _ := inmemory.NewTelemetryStorage()
	logger := logging.NewLogger(nil)
	rec := &EventsRecorderMock{}

	worker := NewEventsWorker(logger, ts, rec, is, &conf.Events{})

	rec.On("Record", []dtos.EventDTO{
		{Key: "key1", TrafficTypeName: "user", EventTypeID: "checkin", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	}, dtos.Metadata{SDKVersion: "php-1.2.3", MachineIP: "", MachineName: ""}).
		Return(nil).
		Once()

	rec.On("Record", []dtos.EventDTO{
		{Key: "key2", TrafficTypeName: "user", EventTypeID: "checkin", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	}, dtos.Metadata{SDKVersion: "go-1.2.3", MachineIP: "", MachineName: ""}).
		Return(nil).
		Once()

	rec.On("Record", []dtos.EventDTO{
		{Key: "key3", TrafficTypeName: "user", EventTypeID: "checkout", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
		{Key: "key4", TrafficTypeName: "user", EventTypeID: "checkout", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	}, dtos.Metadata{SDKVersion: "python-1.2.3", MachineIP: "", MachineName: ""}).Return(nil).Once()

	is.Push(types.ClientMetadata{ID: "i1", SdkVersion: "php-1.2.3"},
		dtos.EventDTO{Key: "key1", TrafficTypeName: "user", EventTypeID: "checkin", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	)
	is.Push(types.ClientMetadata{ID: "i2", SdkVersion: "go-1.2.3"},
		dtos.EventDTO{Key: "key2", TrafficTypeName: "user", EventTypeID: "checkin", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	)

	worker.SynchronizeEvents(5000)
	is.Push(types.ClientMetadata{ID: "i3", SdkVersion: "python-1.2.3"},
		dtos.EventDTO{Key: "key3", TrafficTypeName: "user", EventTypeID: "checkout", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
		dtos.EventDTO{Key: "key4", TrafficTypeName: "user", EventTypeID: "checkout", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	)

	worker.SynchronizeEvents(5000)

	rec.AssertExpectations(t)
}

func TestEventsTaskNoParallelism(t *testing.T) {

	// to test this, we set up a Recorder that sleeps for 1 second and returns (no err).
	// we one call to `SyncrhonizeImpressions()` wait for 500ms, and fire another one.
	// the second one should finish immediately, (becase it does nothing). The second one
	// should finish after 2 seconds

	es, _ := sss.NewEventsQueue(100)
	ts, _ := inmemory.NewTelemetryStorage()
	logger := logging.NewLogger(nil)
	rec := &EventsRecorderMock{}

	worker := NewEventsWorker(logger, ts, rec, es, &conf.Events{})

	rec.On("Record", mock.Anything, mock.Anything).Run(func(mock.Arguments) { time.Sleep(1 * time.Second) }).Return(nil).Twice()

	es.Push(types.ClientMetadata{ID: "i1", SdkVersion: "php-1.2.3"},
		dtos.EventDTO{Key: "key1", TrafficTypeName: "user", EventTypeID: "checkout", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	)
	es.Push(types.ClientMetadata{ID: "i2", SdkVersion: "go-1.2.3"},
		dtos.EventDTO{Key: "key2", TrafficTypeName: "user", EventTypeID: "checkout", Value: nil, Timestamp: 123, Properties: map[string]interface{}{"a": 2}},
	)

	done := make(chan struct{})

	go func() {
		worker.SynchronizeEvents(5000)
		done <- struct{}{}
	}()

	time.Sleep(500 * time.Millisecond)
	assert.Nil(t, worker.SynchronizeEvents(5000))

	// 2nd call has finished, assert that the first one hasn't:
	select {
	case <-done: // first call has finished, fail the test
		assert.Fail(t, "first call shouldn't have finished yet")
	default:
	}

	<-done // blocking wait for 1st to finish

}

type EventsRecorderMock struct {
	mock.Mock
}

// Record implements service.EventsRecorder
func (m *EventsRecorderMock) Record(events []dtos.EventDTO, metadata dtos.Metadata) error {
	args := m.Called(events, metadata)
	return args.Error(0)
}

var _ service.EventsRecorder = (*EventsRecorderMock)(nil)
