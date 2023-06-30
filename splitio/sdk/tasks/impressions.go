package tasks

import (
	"errors"

	sdkconf "github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/types"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-split-commons/v4/storage"
	"github.com/splitio/go-toolkit/v5/asynctask"
	"github.com/splitio/go-toolkit/v5/logging"
	sss "github.com/splitio/splitd/splitio/sdk/storage"
)

type impressionSyncTaskHelper struct {
	logger    logging.LoggerInterface
	telemetry storage.TelemetryRuntimeProducer
	llrec     service.ImpressionsRecorder
	iq        *sss.ImpressionsStorage
	cfg       *sdkconf.Impressions
}

func NewImpressionSyncTask(
	llrec service.ImpressionsRecorder,
	impStore *sss.ImpressionsStorage,
	logger logging.LoggerInterface,
	telemetry storage.TelemetryRuntimeProducer,
	cfg *sdkconf.Impressions,
) *asynctask.AsyncTask {

	helper := &impressionSyncTaskHelper{
		logger:    logger,
		telemetry: telemetry,
		llrec:     llrec,
		iq:        impStore,
		cfg:       cfg,
	}

	return asynctask.NewAsyncTask(
		"impressions-sender",
		func(logging.LoggerInterface) error { helper.synchronize(cfg.PostConcurrency); return nil },
		int(cfg.SyncPeriod.Seconds()),
		nil,
		func(logging.LoggerInterface) { helper.synchronize(cfg.PostConcurrency) },
		logger,
	)

}

func (i *impressionSyncTaskHelper) synchronize(parallelism int) []error {

	var errs []error
	if err := i.iq.RangeAndClear(func(md types.ClientMetadata, q *sss.LockingQueue[dtos.Impression]) {
		extracted := make([]dtos.Impression, 0, q.Len())
		n, err := q.Pop(q.Len(), &extracted)
		if err != nil && !errors.Is(err, sss.ErrQueueEmpty) {
			i.logger.Error("error fetching items from queue: ", err)
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

		if err := i.llrec.Record(payload, dtos.Metadata{SDKVersion: md.SdkVersion}, nil); err != nil {
			errs = append(errs, err)
		}
	}); err != nil {
		i.logger.Error("error traversing impression queues: ", err)
	}

	return errs
}
