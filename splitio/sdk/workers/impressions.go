package workers

import (
	"errors"
	"sync"

	sdkconf "github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"
	serrors "github.com/splitio/splitd/splitio/util/errors"

	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/go-split-commons/v6/service"
	"github.com/splitio/go-split-commons/v6/storage"
	"github.com/splitio/go-split-commons/v6/synchronizer/worker/impression"
	"github.com/splitio/go-toolkit/v5/logging"
	gtsync "github.com/splitio/go-toolkit/v5/sync"
)

type MultiMetaImpressionWorker struct {
	logger    logging.LoggerInterface
	telemetry storage.TelemetryRuntimeProducer
	llrec     service.ImpressionsRecorder
	iq        *sss.ImpressionsStorage
	cfg       *sdkconf.Impressions
	runnning  gtsync.AtomicBool
}

func NewImpressionsWorker(
	logger logging.LoggerInterface,
	telemetry storage.TelemetryRuntimeProducer,
	llrec service.ImpressionsRecorder,
	iq *sss.ImpressionsStorage,
	cfg *sdkconf.Impressions,
) *MultiMetaImpressionWorker {
	return &MultiMetaImpressionWorker{
		logger:    logger,
		telemetry: telemetry,
		llrec:     llrec,
		iq:        iq,
		cfg:       cfg,
	}
}

// FlushImpressions implements impression.ImpressionRecorder
// TODO(mredolatti): take `bulkSize` into account
func (m *MultiMetaImpressionWorker) FlushImpressions(bulkSize int64) error {

	// prevent 2 evictions from happening at the same time. we don't want a sync.Mutex since that would only cause 43928729
	// function calls to pile up and get called after each mutex release.
	if !m.runnning.TestAndSet() {
		m.logger.Warning("flush/sync requested while another one is in progress. ignoring")
		return nil
	}
	defer m.runnning.Unset()

	var errs serrors.ConcurrentErrorCollector
	var wg sync.WaitGroup

	// iterate all internal queues (one per thin-client associate-data)
	// for each [metadata, impressions] tuple, format impressions accordingly, and create a goroutine to post them in BG.
	// after all impressions posting-goroutines have been created, wait for all of them to complete, collect errors,
	// and unset the `running` flag so that this func can be called again
	if err := m.iq.RangeAndClear(func(md types.ClientMetadata, q *sss.LockingQueue[dtos.Impression]) {
		extracted := make([]dtos.Impression, 0, q.Len())
		n, err := q.Pop(q.Len(), &extracted)
		if err != nil && !errors.Is(err, sss.ErrQueueEmpty) {
			m.logger.Error("error fetching items from queue: ", err)
			return // continue with next one
		}

		if n == 0 {
			return // nothing to do here
		}

		formatted := formatImpressions(extracted)

		wg.Add(1)
		go func(imps []dtos.ImpressionsDTO, md dtos.Metadata) {
			defer wg.Done()
			if err := m.llrec.Record(imps, md, nil); err != nil {
				errs.Append(err)
			}
		}(formatted, dtos.Metadata{SDKVersion: md.SdkVersion})
	}); err != nil {
		m.logger.Error("error traversing impression queues: ", err)
	}

	wg.Wait()
	return errs.Join()
}

// SynchronizeImpressions implements impression.ImpressionRecorder
func (m *MultiMetaImpressionWorker) SynchronizeImpressions(bulkSize int64) error {
	return m.FlushImpressions(bulkSize)
}

func formatImpressions(imps []dtos.Impression) []dtos.ImpressionsDTO {
	tmp := make(map[string]*dtos.ImpressionsDTO)
	for idx := range imps {
		forFeature, ok := tmp[imps[idx].FeatureName]
		if !ok {
			forFeature = &dtos.ImpressionsDTO{TestName: imps[idx].FeatureName}
			tmp[imps[idx].FeatureName] = forFeature
		}

		forFeature.KeyImpressions = append(forFeature.KeyImpressions, dtos.ImpressionDTO{
			KeyName:      imps[idx].KeyName,
			Treatment:    imps[idx].Treatment,
			Time:         imps[idx].Time,
			ChangeNumber: imps[idx].ChangeNumber,
			Label:        imps[idx].Label,
			BucketingKey: imps[idx].BucketingKey,
			Pt:           imps[idx].Pt,
		})
	}

	formatted := make([]dtos.ImpressionsDTO, 0, len(tmp))
	for _, v := range tmp {
		formatted = append(formatted, *v)
	}

	return formatted
}

var _ impression.ImpressionRecorder = (*MultiMetaImpressionWorker)(nil)
