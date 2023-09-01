package workers

import (
	"reflect"
	"sort"
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

func TestImpressionsTask(t *testing.T) {
	is, _ := sss.NewImpressionsQueue(100)
	ts, _ := inmemory.NewTelemetryStorage()
	logger := logging.NewLogger(nil)
	rec := &RecorderMock{}

	worker := NewImpressionsWorker(logger, ts, rec, is, &conf.Impressions{})

	var emptyMap map[string]string
	rec.On("Record", []dtos.ImpressionsDTO{{
		TestName:       "f1",
		KeyImpressions: []dtos.ImpressionDTO{{KeyName: "k1", Treatment: "on", Time: 123456, ChangeNumber: 123, Label: "l1"}},
	}}, dtos.Metadata{SDKVersion: "php-1.2.3", MachineIP: "", MachineName: ""}, emptyMap).
		Return(nil).
		Once()

	rec.On("Record", []dtos.ImpressionsDTO{{
		TestName:       "f2",
		KeyImpressions: []dtos.ImpressionDTO{{KeyName: "k2", Treatment: "off", Time: 123457, ChangeNumber: 456, Label: "l2"}},
	}}, dtos.Metadata{SDKVersion: "go-1.2.3", MachineIP: "", MachineName: ""}, emptyMap).
		Return(nil).
		Once()

		// ImpressionsDTO are built from the contents of a map so the ordering is undefined.
		// to solve this, we sort the input by feature name (& provided an already sorted expected value)
	rec.On("Record",
		mock.MatchedBy(func(imps []dtos.ImpressionsDTO) bool {
			sort.Slice(imps, func(i, j int) bool { return imps[i].TestName < imps[j].TestName })
			return reflect.DeepEqual(imps, []dtos.ImpressionsDTO{
				{
					TestName:       "f3",
					KeyImpressions: []dtos.ImpressionDTO{{KeyName: "k3", Treatment: "on", Time: 123458, ChangeNumber: 789, Label: "l3"}},
				},
				{
					TestName:       "f4",
					KeyImpressions: []dtos.ImpressionDTO{{KeyName: "k3", Treatment: "on", Time: 123459, ChangeNumber: 890, Label: "l4"}},
				},
			})
		}),
		dtos.Metadata{SDKVersion: "python-1.2.3", MachineIP: "", MachineName: ""},
		emptyMap).Return(nil).Once()

	is.Push(types.ClientMetadata{ID: "i1", SdkVersion: "php-1.2.3"},
		dtos.Impression{KeyName: "k1", FeatureName: "f1", Treatment: "on", Label: "l1", ChangeNumber: 123, Time: 123456})
	is.Push(types.ClientMetadata{ID: "i2", SdkVersion: "go-1.2.3"},
		dtos.Impression{KeyName: "k2", FeatureName: "f2", Treatment: "off", Label: "l2", ChangeNumber: 456, Time: 123457})

	worker.SynchronizeImpressions(5000)
	is.Push(types.ClientMetadata{ID: "i3", SdkVersion: "python-1.2.3"},
		dtos.Impression{KeyName: "k3", FeatureName: "f3", Treatment: "on", Label: "l3", ChangeNumber: 789, Time: 123458},
		dtos.Impression{KeyName: "k3", FeatureName: "f4", Treatment: "on", Label: "l4", ChangeNumber: 890, Time: 123459},
	)

	worker.SynchronizeImpressions(5000)

	rec.AssertExpectations(t)
}

func TestImpressionsTaskNoParallelism(t *testing.T) {

	// to test this, we set up a Recorder that sleeps for 1 second and returns (no err).
	// we one call to `SyncrhonizeImpressions()` wait for 500ms, and fire another one.
	// the second one should finish immediately, (becase it does nothing). The second one
	// should finish after 2 seconds

	is, _ := sss.NewImpressionsQueue(100)
	ts, _ := inmemory.NewTelemetryStorage()
	logger := logging.NewLogger(nil)
	rec := &RecorderMock{}

	worker := NewImpressionsWorker(logger, ts, rec, is, &conf.Impressions{})

	rec.On("Record", mock.Anything, mock.Anything, mock.Anything).Run(func(mock.Arguments) { time.Sleep(1 * time.Second) }).Return(nil).Twice()

	is.Push(types.ClientMetadata{ID: "i1", SdkVersion: "php-1.2.3"},
		dtos.Impression{KeyName: "k1", FeatureName: "f1", Treatment: "on", Label: "l1", ChangeNumber: 123, Time: 123456})
	is.Push(types.ClientMetadata{ID: "i2", SdkVersion: "go-1.2.3"},
		dtos.Impression{KeyName: "k2", FeatureName: "f2", Treatment: "off", Label: "l2", ChangeNumber: 456, Time: 123457})

	done := make(chan struct{})

	go func() {
		worker.SynchronizeImpressions(5000)
		done <- struct{}{}
	}()

	time.Sleep(500 * time.Millisecond)
	assert.Nil(t, worker.SynchronizeImpressions(5000))

	// 2nd call has finished, assert that the first one hasn't:
	select {
	case <-done: // first call has finished, fail the test
		assert.Fail(t, "first call shouldn't have finished yet")
	default:
	}

	<-done // blocking wait for 1st to finish

}

type RecorderMock struct {
	mock.Mock
}

// Record implements service.ImpressionsRecorder
func (m *RecorderMock) Record(impressions []dtos.ImpressionsDTO, metadata dtos.Metadata, extraHeaders map[string]string) error {
	args := m.Called(impressions, metadata, extraHeaders)
	return args.Error(0)
}

// RecordImpressionsCount implements service.ImpressionsRecorder
func (m *RecorderMock) RecordImpressionsCount(pf dtos.ImpressionsCountDTO, metadata dtos.Metadata) error {
	args := m.Called(pf, metadata)
	return args.Error(0)
}

var _ service.ImpressionsRecorder = (*RecorderMock)(nil)
