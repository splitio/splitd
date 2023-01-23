package sdk

import (
	"fmt"

	sss "github.com/splitio/splitd/splitio/sdk/storage"
	tss "github.com/splitio/splitd/splitio/sdk/tasks"
	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"

	"github.com/splitio/go-split-commons/v4/conf"
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/provisional"
	"github.com/splitio/go-split-commons/v4/provisional/strategy"
	"github.com/splitio/go-split-commons/v4/service/api"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/storage/inmemory"
	"github.com/splitio/go-split-commons/v4/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v4/synchronizer"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/impressionscount"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/segment"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v4/tasks"

	storageCommon "github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/logging"
)

func setupWorkers(logger logging.LoggerInterface, api *api.SplitAPI, str *storages) (*synchronizer.Workers, error) {
	hc := &application.Dummy{}
	splitChangeWorker := split.NewSplitFetcher(str.splits, api.SplitFetcher, logger, str.telemetry, hc)
	segmentChangeWorker := segment.NewSegmentFetcher(str.splits, str.segments, api.SegmentFetcher, logger, str.telemetry, hc)
	return &synchronizer.Workers{
		SplitFetcher:   splitChangeWorker,
		SegmentFetcher: segmentChangeWorker,
	}, nil
}

func setupTasks(
	cfg *sdkConf.Config,
	str *storages,
	logger logging.LoggerInterface,
	workers *synchronizer.Workers,
	impComponents impComponents,
	md dtos.Metadata,
	api *api.SplitAPI,
) (*synchronizer.SplitTasks, error) {

	impCfg := cfg.Impressions

	return &synchronizer.SplitTasks{
		SplitSyncTask:      tasks.NewFetchSplitsTask(workers.SplitFetcher, int(cfg.Splits.SyncPeriod.Seconds()), logger),
		SegmentSyncTask:    tasks.NewFetchSegmentsTask(workers.SegmentFetcher, int(cfg.Segments.SyncPeriod.Seconds()), cfg.Segments.WorkerCount, cfg.Segments.QueueSize, logger),
		ImpressionSyncTask: tss.NewImpressionSyncTask(api.ImpressionRecorder, str.impressions, logger, str.telemetry, &cfg.Impressions),
		ImpressionsCountSyncTask: tasks.NewRecordImpressionsCountTask(
			impressionscount.NewRecorderSingle(impComponents.counter, api.ImpressionRecorder, md, logger, str.telemetry),
			logger,
			int(impCfg.CountSyncPeriod.Seconds()),
		),
		TelemetrySyncTask:     &NoOpTask{},
		EventSyncTask:         &NoOpTask{},
		UniqueKeysTask:        &NoOpTask{},
		CleanFilterTask:       &NoOpTask{},
		ImpsCountConsumerTask: &NoOpTask{},
	}, nil
}

type impComponents struct {
	manager provisional.ImpressionManager
	counter *strategy.ImpressionsCounter
}

func setupImpressionsComponents(c *sdkConf.Impressions, telemetry storageCommon.TelemetryRuntimeProducer) (impComponents, error) {

	observer, err := strategy.NewImpressionObserver(c.ObserverSize)
	if err != nil {
		return impComponents{}, fmt.Errorf("error building impressions observer: %w", err)
	}

	var s strategy.ProcessStrategyInterface
	var counter *strategy.ImpressionsCounter
	switch c.Mode {
	case conf.ImpressionsModeDebug:
		s = strategy.NewDebugImpl(observer, false)
	case conf.ImpressionsModeNone:
	default: // optimized
		counter = strategy.NewImpressionsCounter()
		s = strategy.NewOptimizedImpl(observer, counter, telemetry, false)
	}

	return impComponents{
		manager: provisional.NewImpressionManager(s),
		counter: counter,
	}, nil
}

type storages struct {
	splits      storage.SplitStorage
	segments    storage.SegmentStorage
	telemetry   storage.TelemetryStorage
	impressions *sss.ImpressionsStorage
}

func setupStorages(cfg *sdkConf.Config) *storages {
	ts, _ := inmemory.NewTelemetryStorage()
	iq, _ := sss.NewImpressionsQueue(cfg.Impressions.QueueSize)

	return &storages{
		splits:      mutexmap.NewMMSplitStorage(),
		segments:    mutexmap.NewMMSegmentStorage(),
		impressions: iq,
		telemetry:   ts,
	}
}

// Temporary for running without impressions/events/telemetry/etc
type NoOpTask struct{}

func (*NoOpTask) IsRunning() bool          { return false }
func (*NoOpTask) Start()                   {}
func (*NoOpTask) Stop(blocking bool) error { return nil }

var _ tasks.Task = (*NoOpTask)(nil)
