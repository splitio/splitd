package sdk

import (
	"fmt"

	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/workers"

	"github.com/splitio/go-split-commons/v6/conf"
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/flagsets"
	"github.com/splitio/go-split-commons/v6/healthcheck/application"
	"github.com/splitio/go-split-commons/v6/provisional"
	"github.com/splitio/go-split-commons/v6/provisional/strategy"
	"github.com/splitio/go-split-commons/v6/service/api"
	"github.com/splitio/go-split-commons/v6/storage"
	"github.com/splitio/go-split-commons/v6/storage/filter"
	"github.com/splitio/go-split-commons/v6/storage/inmemory"
	"github.com/splitio/go-split-commons/v6/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v6/synchronizer"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/impressionscount"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/segment"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v6/tasks"
	"github.com/splitio/go-split-commons/v6/telemetry"
	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	bfExpectedElemenets                = 10000000
	bfFalsePositiveProbability         = 0.01
	bfCleaningPeriod                   = 86400 // 24 hours
	uniqueKeysPeriodTaskInMemory       = 900   // 15 min
	uniqueKeysPeriodTaskRedis          = 300   // 5 min
	impressionsCountPeriodTaskInMemory = 1800  // 30 min
	impressionsCountPeriodTaskRedis    = 300   // 5 min
	impressionsBulkSizeRedis           = 100
)

func setupWorkers(
	logger logging.LoggerInterface,
	api *api.SplitAPI,
	str *storages,
	hc application.MonitorProducerInterface,
	cfg *sdkConf.Config,
	flagSetsFilter flagsets.FlagSetFilter,
	md dtos.Metadata,
	impComponents impComponents,
) *synchronizer.Workers {
	return &synchronizer.Workers{
		SplitUpdater:             split.NewSplitUpdater(str.splits, api.SplitFetcher, logger, str.telemetry, hc, flagSetsFilter),
		SegmentUpdater:           segment.NewSegmentUpdater(str.splits, str.segments, api.SegmentFetcher, logger, str.telemetry, hc),
		ImpressionRecorder:       workers.NewImpressionsWorker(logger, str.telemetry, api.ImpressionRecorder, str.impressions, &cfg.Impressions),
		EventRecorder:            workers.NewEventsWorker(logger, str.telemetry, api.EventRecorder, str.events, &cfg.Events),
		ImpressionsCountRecorder: impressionscount.NewRecorderSingle(impComponents.counter, api.ImpressionRecorder, md, logger, str.telemetry),
		TelemetryRecorder:        telemetry.NewTelemetrySynchronizer(str.telemetry, api.TelemetryRecorder, str.splits, str.segments, logger, md, str.telemetry),
	}
}

func setupTasks(
	cfg *sdkConf.Config,
	logger logging.LoggerInterface,
	workers *synchronizer.Workers,
	impComponents impComponents,
) *synchronizer.SplitTasks {
	impCfg := cfg.Impressions
	evCfg := cfg.Events
	dummyHC := &application.Dummy{}
	tg := &synchronizer.SplitTasks{
		SplitSyncTask: tasks.NewFetchSplitsTask(workers.SplitUpdater, int(cfg.Splits.SyncPeriod.Seconds()), logger),
		SegmentSyncTask: tasks.NewFetchSegmentsTask(
			workers.SegmentUpdater,
			int(cfg.Segments.SyncPeriod.Seconds()),
			cfg.Segments.WorkerCount,
			cfg.Segments.QueueSize,
			logger,
			dummyHC,
		),
		ImpressionSyncTask: tasks.NewRecordImpressionsTask(workers.ImpressionRecorder, int(impCfg.SyncPeriod.Seconds()), logger, 5000),
		EventSyncTask:      tasks.NewRecordEventsTask(workers.EventRecorder, 5000, int(evCfg.SyncPeriod.Seconds()), logger),
		TelemetrySyncTask:  &NoOpTask{},
		UniqueKeysTask:     tasks.NewRecordUniqueKeysTask(workers.TelemetryRecorder, *impComponents.tracker, uniqueKeysPeriodTaskInMemory, logger),
		CleanFilterTask:    tasks.NewCleanFilterTask(*impComponents.filter, logger, bfCleaningPeriod),
		ImpsCountConsumerTask: tasks.NewRecordImpressionsCountTask(
			workers.ImpressionsCountRecorder,
			logger,
			int(impCfg.CountSyncPeriod.Seconds()),
		),
	}

	return tg
}

type impComponents struct {
	manager provisional.ImpressionManager
	counter *strategy.ImpressionsCounter
	tracker *strategy.UniqueKeysTracker
	filter  *storage.Filter
}

func setupImpressionsComponents(c *sdkConf.Impressions, telemetry storage.TelemetryRuntimeProducer) (impComponents, error) {

	observer, err := strategy.NewImpressionObserver(c.ObserverSize)
	if err != nil {
		return impComponents{}, fmt.Errorf("error building impressions observer: %w", err)
	}

	counter := strategy.NewImpressionsCounter()
	bf := filter.NewBloomFilter(bfExpectedElemenets, bfFalsePositiveProbability)
	tracker := strategy.NewUniqueKeysTracker(bf)
	none := strategy.NewNoneImpl(counter, tracker, false)

	var s strategy.ProcessStrategyInterface
	switch c.Mode {
	case conf.ImpressionsModeDebug:
		s = strategy.NewDebugImpl(observer, false)
	case conf.ImpressionsModeNone:
	default: // optimized
		s = strategy.NewOptimizedImpl(observer, counter, telemetry, false)
	}

	impManager := provisional.NewImpressionManagerImp(none, s)

	return impComponents{
		manager: impManager,
		counter: counter,
		tracker: &tracker,
		filter:  &bf,
	}, nil
}

type storages struct {
	splits      storage.SplitStorage
	segments    storage.SegmentStorage
	telemetry   storage.TelemetryStorage
	impressions *sss.ImpressionsStorage
	events      *sss.EventsStorage
}

func setupStorages(cfg *sdkConf.Config, flagSetsFilter flagsets.FlagSetFilter) *storages {
	ts, _ := inmemory.NewTelemetryStorage()
	iq, _ := sss.NewImpressionsQueue(cfg.Impressions.QueueSize)
	eq, _ := sss.NewEventsQueue(cfg.Events.QueueSize)

	return &storages{
		splits:      mutexmap.NewMMSplitStorage(flagSetsFilter),
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
