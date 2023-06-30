package tasks

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
	"github.com/stretchr/testify/mock"
)

func TestImpressionsTask(t *testing.T) {
	is, _ := sss.NewImpressionsQueue(100)
	ts, _ := inmemory.NewTelemetryStorage()
	logger := logging.NewLogger(nil)
	rec := &RecorderMock{}

	task := NewImpressionSyncTask(rec, is, logger, ts, &conf.Impressions{SyncPeriod: 1 * time.Second})

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

	rec.On("Record", []dtos.ImpressionsDTO{
		{
			TestName:       "f3",
			KeyImpressions: []dtos.ImpressionDTO{{KeyName: "k3", Treatment: "on", Time: 123458, ChangeNumber: 789, Label: "l3"}},
		},
		{
			TestName:       "f4",
			KeyImpressions: []dtos.ImpressionDTO{{KeyName: "k3", Treatment: "on", Time: 123459, ChangeNumber: 890, Label: "l4"}},
		},
	}, dtos.Metadata{SDKVersion: "python-1.2.3", MachineIP: "", MachineName: ""}, emptyMap).
		Return(nil).
		Once()

	is.Push(types.ClientMetadata{ID: "i1", SdkVersion: "php-1.2.3"},
		dtos.Impression{KeyName: "k1", FeatureName: "f1", Treatment: "on", Label: "l1", ChangeNumber: 123, Time: 123456})
	is.Push(types.ClientMetadata{ID: "i2", SdkVersion: "go-1.2.3"},
		dtos.Impression{KeyName: "k2", FeatureName: "f2", Treatment: "off", Label: "l2", ChangeNumber: 456, Time: 123457})

	task.Start()
	time.Sleep(1500 * time.Millisecond)

	is.Push(types.ClientMetadata{ID: "i3", SdkVersion: "python-1.2.3"},
		dtos.Impression{KeyName: "k3", FeatureName: "f3", Treatment: "on", Label: "l3", ChangeNumber: 789, Time: 123458},
		dtos.Impression{KeyName: "k3", FeatureName: "f4", Treatment: "on", Label: "l4", ChangeNumber: 890, Time: 123459},
	)

	time.Sleep(1500 * time.Millisecond)
	task.Stop(true)

    rec.AssertExpectations(t)
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
