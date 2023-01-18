package sdk

import (
	"github.com/splitio/go-split-commons/v4/conf"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/service/api"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/synchronizer"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/segment"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v4/tasks"
	"github.com/splitio/go-toolkit/v5/logging"
)

type storages struct {
	splits    storage.SplitStorage
	segments  storage.SegmentStorage
	telemetry storage.TelemetryStorage
}

func setupWorkers(logger logging.LoggerInterface, api *api.SplitAPI, str *storages) (*synchronizer.Workers, error) {
	hc := &application.Dummy{}
	splitChangeWorker := split.NewSplitFetcher(str.splits, api.SplitFetcher, logger, str.telemetry, hc)
	segmentChangeWorker := segment.NewSegmentFetcher(str.splits, str.segments, api.SegmentFetcher, logger, str.telemetry, hc)
	return &synchronizer.Workers{
		SplitFetcher:   splitChangeWorker,
		SegmentFetcher: segmentChangeWorker,
	}, nil
}

func setupTastsk(cfg *conf.AdvancedConfig, logger logging.LoggerInterface, workers *synchronizer.Workers) (*synchronizer.SplitTasks, error) {
	return &synchronizer.SplitTasks{
		SplitSyncTask: tasks.NewFetchSplitsTask(workers.SplitFetcher, cfg.SplitsRefreshRate, logger),
		SegmentSyncTask: tasks.NewFetchSegmentsTask(workers.SegmentFetcher, cfg.SegmentsRefreshRate, cfg.SegmentWorkers, cfg.SegmentQueueSize, logger),
		TelemetrySyncTask: &NoOpTask{},
		ImpressionSyncTask: &NoOpTask{},
		EventSyncTask: &NoOpTask{},
		ImpressionsCountSyncTask: &NoOpTask{},
		UniqueKeysTask: &NoOpTask{},
		CleanFilterTask: &NoOpTask{},
		ImpsCountConsumerTask: &NoOpTask{},
	}, nil
}


// Temporary for running without impressions/events/telemetry/etc
type NoOpTask struct{}

func (*NoOpTask) IsRunning() bool          { return false }
func (*NoOpTask) Start()                   {}
func (*NoOpTask) Stop(blocking bool) error { return nil }

var _ tasks.Task = (*NoOpTask)(nil)
