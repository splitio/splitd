package workers

import (
	"errors"
	"sync"

	sdkconf "github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"
	serrors "github.com/splitio/splitd/splitio/util/errors"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/event"
	"github.com/splitio/go-toolkit/v5/logging"
	gtsync "github.com/splitio/go-toolkit/v5/sync"
)

type MultiMetaEventsWorker struct {
	logger    logging.LoggerInterface
	telemetry storage.TelemetryRuntimeProducer
	llrec     service.EventsRecorder
	iq        *sss.EventsStorage
	cfg       *sdkconf.Events
	runnning  gtsync.AtomicBool
}

func NewEventsWorker(
	logger logging.LoggerInterface,
	telemetry storage.TelemetryRuntimeProducer,
	llrec service.EventsRecorder,
	iq *sss.EventsStorage,
	cfg *sdkconf.Events,
) *MultiMetaEventsWorker {
	return &MultiMetaEventsWorker{
		logger:    logger,
		telemetry: telemetry,
		llrec:     llrec,
		iq:        iq,
		cfg:       cfg,
	}
}

// FlushImpressions implements impression.ImpressionRecorder
// TODO(mredolatti): take `bulkSize` into account
func (m *MultiMetaEventsWorker) FlushEvents(bulkSize int64) error {

	// prevent 2 evictions from happening at the same time. we don't want a sync.Mutex since that would only cause 43928729
	// function calls to pile up and get called after each mutex release.
	if !m.runnning.TestAndSet() {
		m.logger.Warning("flush/sync requested while another one is in progress. ignoring")
		return nil
	}
	defer m.runnning.Unset()

	var errs serrors.ConcurrentErrorCollector
	var wg sync.WaitGroup

    // same logic as impressions workers, without the need for formatting. check impressions.go for a better
    // description of what's being done
    if err := m.iq.RangeAndClear(func(md types.ClientMetadata, q *sss.LockingQueue[dtos.EventDTO]) {
		extracted := make([]dtos.EventDTO, 0, q.Len())
		n, err := q.Pop(q.Len(), &extracted)
		if err != nil && !errors.Is(err, sss.ErrQueueEmpty) {
			m.logger.Error("error fetching items from queue: ", err)
			return // continue with queue
		}

		if n == 0 {
			return // nothing to do here
		}

		wg.Add(1)
		go func(events []dtos.EventDTO, cc types.ClientMetadata) {
			defer wg.Done()
			if err := m.llrec.Record(events, dtos.Metadata{SDKVersion: cc.SdkVersion}); err != nil {
				errs.Append(err)
			}
		}(extracted, md)
	}); err != nil {
		m.logger.Error("error traversing event queues: ", err)
	}

	wg.Wait()
	return errs.Join()
}

// SynchronizeImpressions implements impression.ImpressionRecorder
func (m *MultiMetaEventsWorker) SynchronizeEvents(bulkSize int64) error {
	return m.FlushEvents(bulkSize)
}

var _ event.EventRecorder = (*MultiMetaEventsWorker)(nil)
