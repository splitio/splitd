package workers

import (
	"errors"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-split-commons/v4/synchronizer/worker/impression"
	"github.com/splitio/go-toolkit/v5/logging"
	sdkconf "github.com/splitio/splitd/splitio/sdk/conf"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"
)

type MultiMetaImpressionWorker struct {
	logger    logging.LoggerInterface
	telemetry storage.TelemetryRuntimeProducer
	llrec     service.ImpressionsRecorder
	iq        *sss.ImpressionsStorage
	cfg       *sdkconf.Impressions
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
func (m *MultiMetaImpressionWorker) FlushImpressions(bulkSize int64) error {

    // TODO(mredolatti): take `bulkSize` into account
	var errs []error
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

		tmp := make(map[string]*dtos.ImpressionsDTO)
		for idx := range extracted {
			forFeature, ok := tmp[extracted[idx].FeatureName]
			if !ok {
				forFeature = &dtos.ImpressionsDTO{TestName: extracted[idx].FeatureName}
				tmp[extracted[idx].FeatureName] = forFeature
			}

			forFeature.KeyImpressions = append(forFeature.KeyImpressions, dtos.ImpressionDTO{
				KeyName:      extracted[idx].KeyName,
				Treatment:    extracted[idx].Treatment,
				Time:         extracted[idx].Time,
				ChangeNumber: extracted[idx].ChangeNumber,
				Label:        extracted[idx].Label,
				BucketingKey: extracted[idx].BucketingKey,
				Pt:           extracted[idx].Pt,
			})
		}

		payload := make([]dtos.ImpressionsDTO, 0, len(tmp))
		for _, v := range tmp {
			payload = append(payload, *v)
		}

		if err := m.llrec.Record(payload, dtos.Metadata{SDKVersion: md.SdkVersion}, nil); err != nil {
			errs = append(errs, err)
		}
	}); err != nil {
		m.logger.Error("error traversing impression queues: ", err)
	}

	return errors.Join(errs...)
}

// SynchronizeImpressions implements impression.ImpressionRecorder
func (m *MultiMetaImpressionWorker) SynchronizeImpressions(bulkSize int64) error {
	return m.FlushImpressions(bulkSize)
}

var _ impression.ImpressionRecorder = (*MultiMetaImpressionWorker)(nil)
