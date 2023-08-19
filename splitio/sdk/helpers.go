package sdk

import (
	"fmt"

	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/workers"

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

func setupWorkers(
	logger logging.LoggerInterface,
	api *api.SplitAPI,
	str *storages,
	hc application.MonitorProducerInterface,
	cfg *sdkConf.Config,

) *synchronizer.Workers {
	return &synchronizer.Workers{
		SplitFetcher:       split.NewSplitFetcher(str.splits, api.SplitFetcher, logger, str.telemetry, hc),
		SegmentFetcher:     segment.NewSegmentFetcher(str.splits, str.segments, api.SegmentFetcher, logger, str.telemetry, hc),
		ImpressionRecorder: workers.NewImpressionsWorker(logger, str.telemetry, api.ImpressionRecorder, str.impressions, &cfg.Impressions),
		EventRecorder:      workers.NewEventsWorker(logger, str.telemetry, api.EventRecorder, str.events, &cfg.Events),
	}
}

func setupTasks(
	cfg *sdkConf.Config,
	str *storages,
	logger logging.LoggerInterface,
	workers *synchronizer.Workers,
	impComponents impComponents,
	md dtos.Metadata,
	api *api.SplitAPI,
) *synchronizer.SplitTasks {
	impCfg := cfg.Impressions
	evCfg := cfg.Events
	tg := &synchronizer.SplitTasks{
		SplitSyncTask: tasks.NewFetchSplitsTask(workers.SplitFetcher, int(cfg.Splits.SyncPeriod.Seconds()), logger),
		SegmentSyncTask: tasks.NewFetchSegmentsTask(
			workers.SegmentFetcher,
			int(cfg.Segments.SyncPeriod.Seconds()),
			cfg.Segments.WorkerCount,
			cfg.Segments.QueueSize,
			logger,
		),
		ImpressionSyncTask:    tasks.NewRecordImpressionsTask(workers.ImpressionRecorder, int(impCfg.SyncPeriod.Seconds()), logger, 5000),
		EventSyncTask:         tasks.NewRecordEventsTask(workers.EventRecorder, 5000, int(evCfg.SyncPeriod.Seconds()), logger),
		TelemetrySyncTask:     &NoOpTask{},
		UniqueKeysTask:        &NoOpTask{},
		CleanFilterTask:       &NoOpTask{},
		ImpsCountConsumerTask: &NoOpTask{},
	}

	if impCfg.Mode == "optimized" {
		tg.ImpressionsCountSyncTask = tasks.NewRecordImpressionsCountTask(
			impressionscount.NewRecorderSingle(impComponents.counter, api.ImpressionRecorder, md, logger, str.telemetry),
			logger,
			int(impCfg.CountSyncPeriod.Seconds()),
		)
	}

	return tg
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
	events      *sss.EventsStorage
}

func setupStorages(cfg *sdkConf.Config) *storages {
	ts, _ := inmemory.NewTelemetryStorage()
	iq, _ := sss.NewImpressionsQueue(cfg.Impressions.QueueSize)
	eq, _ := sss.NewEventsQueue(cfg.Events.QueueSize)

	return &storages{
		splits:      mutexmap.NewMMSplitStorage(),
		segments:    mutexmap.NewMMSegmentStorage(),
		impressions: iq,
		events:      eq,
		telemetry:   ts,
	}
}

// Temporary for running without impressions/events/telemetry/etc
type NoOpTask struct{}

func (*NoOpTask) IsRunning() bool          { return false }
func (*NoOpTask) Start()                   {}
func (*NoOpTask) Stop(blocking bool) error { return nil }

var _ tasks.Task = (*NoOpTask)(nil)
