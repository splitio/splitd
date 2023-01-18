package sdk

import (
	"fmt"

	"github.com/splitio/go-client/v6/splitio/engine"
	"github.com/splitio/go-client/v6/splitio/engine/evaluator"
	"github.com/splitio/go-split-commons/v4/conf"
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/service/api"
	"github.com/splitio/go-split-commons/v4/storage/inmemory"
	"github.com/splitio/go-split-commons/v4/storage/inmemory/mutexmap"
	"github.com/splitio/go-split-commons/v4/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
)

type ClientMetadata struct {
	ID         string
	SdkVersion string
}

type Interface interface {
	Treatment(md *ClientMetadata, Key string, BucketingKey string, Feature string, attributes map[string]interface{}) (string, error)
}

type Impl struct {
	ev     evaluator.Interface
	sm     synchronizer.Manager
	ss     synchronizer.Synchronizer
	status chan int
}

func New(logger logging.LoggerInterface, apikey string, cfg *conf.AdvancedConfig) (*Impl, error) {

	hc := &application.Dummy{}
	status := make(chan int, 10)
	md := dtos.Metadata{SDKVersion: fmt.Sprintf("splitd-%s", splitio.Version)}
	splitApi := api.NewSplitAPI(apikey, *cfg, logger, md)
	ts, _ := inmemory.NewTelemetryStorage()
	stores := &storages{splits: mutexmap.NewMMSplitStorage(), segments: mutexmap.NewMMSegmentStorage(), telemetry: ts}
	workers, _ := setupWorkers(logger, splitApi, stores)
	tasks, _ := setupTastsk(cfg, logger, workers)
	sync := synchronizer.NewSynchronizer(*cfg, *tasks, *workers, logger, nil, nil)
	manager, err := synchronizer.NewSynchronizerManager(sync, logger, *cfg, splitApi.AuthClient, stores.splits, status, stores.telemetry, md, nil, hc)
	if err != nil {
		return nil, fmt.Errorf("error initializing split evaluation service: %w", err)
	}

	i := &Impl{sm: manager, ss: sync, ev: evaluator.NewEvaluator(stores.splits, stores.segments, engine.NewEngine(logger), logger)}

	i.sm.Start()
	res := <-status
	if res == synchronizer.Error {
		return nil, fmt.Errorf("failed to perform initial sync")
	}

	return i, nil
}

// Treatment implements Interface
func (i *Impl) Treatment(md *ClientMetadata, key string, bucketingKey string, feature string, attributes map[string]interface{}) (string, error) {
	res := i.ev.EvaluateFeature(key, &bucketingKey, feature, attributes)
	if res == nil {
		return "", fmt.Errorf("nil result")
	}
	return res.Treatment, nil
}

var _ Interface = (*Impl)(nil)
