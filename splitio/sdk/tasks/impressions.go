package tasks

import (
	"github.com/splitio/go-toolkit/v5/asynctask"
	"github.com/splitio/go-toolkit/v5/logging"
	sdkconf "github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/workers"
)

const (
    defaultBulkSize = 5000
)

func NewImpressionSyncTask(
	worker *workers.MultiMetaImpressionWorker,
	logger logging.LoggerInterface,
	cfg *sdkconf.Impressions,
) *asynctask.AsyncTask {

	// TODO(mredolatti): pass a proper bulk size (currently ignored, everything is flushed)
	return asynctask.NewAsyncTask(
		"impressions-sender",
		func(logging.LoggerInterface) error { worker.SynchronizeImpressions(defaultBulkSize); return nil },
		int(cfg.SyncPeriod.Seconds()),
		nil,
		func(logging.LoggerInterface) { worker.SynchronizeImpressions(defaultBulkSize) },
		logger,
	)
}
