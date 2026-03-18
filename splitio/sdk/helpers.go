package sdk

import (
	"fmt"
	"regexp"
	"strings"

	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/workers"

	"github.com/splitio/go-split-commons/v9/conf"
	"github.com/splitio/go-split-commons/v9/dtos"
	"github.com/splitio/go-split-commons/v9/engine/grammar"
	"github.com/splitio/go-split-commons/v9/flagsets"
	"github.com/splitio/go-split-commons/v9/healthcheck/application"
	"github.com/splitio/go-split-commons/v9/provisional"
	"github.com/splitio/go-split-commons/v9/provisional/strategy"
	"github.com/splitio/go-split-commons/v9/service/api"
	"github.com/splitio/go-split-commons/v9/service/api/specs"
	"github.com/splitio/go-split-commons/v9/storage"
	"github.com/splitio/go-split-commons/v9/storage/filter"
	"github.com/splitio/go-split-commons/v9/storage/inmemory"
	"github.com/splitio/go-split-commons/v9/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v9/synchronizer"
	"github.com/splitio/go-split-commons/v9/synchronizer/worker/impressionscount"
	"github.com/splitio/go-split-commons/v9/synchronizer/worker/segment"
	"github.com/splitio/go-split-commons/v9/synchronizer/worker/split"
	"github.com/splitio/go-split-commons/v9/tasks"
	"github.com/splitio/go-split-commons/v9/telemetry"
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
	// Max Flag name length
	MaxFlagNameLength = 100
	// Max Treatment length
	MaxTreatmentLength = 100
	// Treatment regexp
	TreatmentRegexp = "^[0-9]+[.a-zA-Z0-9_-]*$|^[a-zA-Z]+[a-zA-Z0-9_-]*$"
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
	ruleBuilder grammar.RuleBuilder,
) *synchronizer.Workers {
	return &synchronizer.Workers{
		SplitUpdater:             split.NewSplitUpdater(str.splits, str.ruleBasedSegments, api.SplitFetcher, logger, str.telemetry, hc, flagSetsFilter, ruleBuilder, false, specs.FLAG_V1_3),
		SegmentUpdater:           segment.NewSegmentUpdater(str.splits, str.segments, str.ruleBasedSegments, api.SegmentFetcher, logger, str.telemetry, hc),
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
		s = none
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
	splits            storage.SplitStorage
	segments          storage.SegmentStorage
	ruleBasedSegments storage.RuleBasedSegmentsStorage
	telemetry         storage.TelemetryStorage
	impressions       *sss.ImpressionsStorage
	events            *sss.EventsStorage
}

func setupStorages(cfg *sdkConf.Config, flagSetsFilter flagsets.FlagSetFilter) *storages {
	ts, _ := inmemory.NewTelemetryStorage()
	iq, _ := sss.NewImpressionsQueue(cfg.Impressions.QueueSize)
	eq, _ := sss.NewEventsQueue(cfg.Events.QueueSize)

	return &storages{
		splits:            mutexmap.NewMMSplitStorage(flagSetsFilter),
		segments:          mutexmap.NewMMSegmentStorage(),
		ruleBasedSegments: mutexmap.NewRuleBasedSegmentsStorage(),
		impressions:       iq,
		events:            eq,
		telemetry:         ts,
	}
}

func SanitizeGlobalFallbackTreatment(global *dtos.FallbackTreatment, logger logging.LoggerInterface) *dtos.FallbackTreatment {
	if global == nil {
		return nil
	}
	if !isValidTreatment(global) {
		logger.Error(fmt.Sprintf("Fallback treatments - Discarded global fallback: Invalid treatment (max %d chars and comply with %s)", MaxTreatmentLength, TreatmentRegexp))
		return nil
	}
	return &dtos.FallbackTreatment{
		Treatment: global.Treatment,
		Config:    global.Config,
	}
}

func isValidTreatment(fallbackTreatment *dtos.FallbackTreatment) bool {
	if fallbackTreatment == nil || fallbackTreatment.Treatment == nil {
		return false
	}
	value := *fallbackTreatment.Treatment
	pattern := regexp.MustCompile(TreatmentRegexp)
	return len(value) <= MaxTreatmentLength && pattern.MatchString(value)
}

func SanitizeByFlagFallBackTreatment(byFlag map[string]dtos.FallbackTreatment, logger logging.LoggerInterface) map[string]dtos.FallbackTreatment {
	sanitized := map[string]dtos.FallbackTreatment{}
	if len(byFlag) == 0 {
		return sanitized
	}
	for flagName, treatment := range byFlag {
		if !isValidFlagName(&flagName) {
			logger.Error(fmt.Sprintf("Fallback treatments - Discarded flag: Invalid flag name (max %d chars, no spaces)", MaxFlagNameLength))
			continue
		}
		if !isValidTreatment(&treatment) {
			logger.Error(fmt.Sprintf("Fallback treatments -  Discarded treatment for flag '%s': Invalid treatment (max %d chars and comply with %s)", flagName, MaxTreatmentLength, TreatmentRegexp))
			continue
		}
		sanitized[flagName] = dtos.FallbackTreatment{
			Treatment: treatment.Treatment,
			Config:    treatment.Config,
		}
	}
	return sanitized
}

func isValidFlagName(flagName *string) bool {
	if flagName == nil {
		return false
	}
	return len(*flagName) <= MaxFlagNameLength && !strings.Contains(*flagName, " ")
}

// Temporary for running without impressions/events/telemetry/etc
type NoOpTask struct{}

func (*NoOpTask) IsRunning() bool          { return false }
func (*NoOpTask) Start()                   {}
func (*NoOpTask) Stop(blocking bool) error { return nil }

var _ tasks.Task = (*NoOpTask)(nil)
