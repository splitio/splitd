package tasks

import (
	"github.com/splitio/go-toolkit/v5/asynctask"
	"github.com/splitio/go-toolkit/v5/logging"
	sdkconf "github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/workers"
)

const (
    defaultEventsBulkSize = 5000
)

func NewEventsSyncTask(
	worker *workers.MultiMetaEventsWorker,
	logger logging.LoggerInterface,
	cfg *sdkconf.Impressions,
) *asynctask.AsyncTask {

	// TODO(mredolatti): pass a proper bulk size (currently ignored, everything is flushed)
	return asynctask.NewAsyncTask(
		"events-sender",
		func(logging.LoggerInterface) error { worker.SynchronizeEvents(defaultEventsBulkSize); return nil },
		int(cfg.SyncPeriod.Seconds()),
		nil,
		func(logging.LoggerInterface) { worker.SynchronizeEvents(defaultEventsBulkSize) },
		logger,
	)
}
