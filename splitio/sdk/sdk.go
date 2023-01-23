package sdk

import (
	"fmt"
	"time"

	"github.com/splitio/splitd/splitio/sdk/conf"
	sdkConf "github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"

	"github.com/splitio/go-client/v6/splitio/engine"
	"github.com/splitio/go-client/v6/splitio/engine/evaluator"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/healthcheck/application"
	"github.com/splitio/go-split-commons/v4/provisional"
	"github.com/splitio/go-split-commons/v4/service/api"
	"github.com/splitio/go-split-commons/v4/synchronizer"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
)

type Interface interface {
	Treatment(md *types.ClientMetadata, Key string, BucketingKey string, Feature string, attributes map[string]interface{}) (string, error)
}

type Impl struct {
	logger logging.LoggerInterface
	ev     evaluator.Interface
	sm     synchronizer.Manager
	ss     synchronizer.Synchronizer
	is     *storage.ImpressionsStorage
	iq     provisional.ImpressionManager
	cfg    sdkConf.Config
	status chan int
}

func New(logger logging.LoggerInterface, apikey string, opts ...conf.Option) (*Impl, error) {

	md := dtos.Metadata{SDKVersion: fmt.Sprintf("splitd-%s", splitio.Version)}
	c := sdkConf.DefaultConfig()
	advCfg := c.ToAdvancedConfig()
	if err := c.ParseOptions(opts); err != nil {
		return nil, fmt.Errorf("error parsing SDK config: %w", err)
	}

	stores := setupStorages(c)
	impc, err := setupImpressionsComponents(&c.Impressions, stores.telemetry)
	if err != nil {
		return nil, fmt.Errorf("error setting up impressions components")
	}

	hc := &application.Dummy{}
	splitApi := api.NewSplitAPI(apikey, *advCfg, logger, md)
	workers := setupWorkers(logger, splitApi, stores, hc)
	tasks := setupTasks(c, stores, logger, workers, impc, md, splitApi)
	sync := synchronizer.NewSynchronizer(*advCfg, *tasks, *workers, logger, nil, nil)

	status := make(chan int, 10)
	manager, err := synchronizer.NewSynchronizerManager(sync, logger, *advCfg, splitApi.AuthClient, stores.splits, status, stores.telemetry, md, nil, hc)
	if err != nil {
		return nil, fmt.Errorf("error initializing split evaluation service: %w", err)
	}

	// Start initial sync and await readiness
	manager.Start()
	res := <-status
	if res == synchronizer.Error {
		return nil, fmt.Errorf("failed to perform initial sync")
	}

	return &Impl{
		logger: logger,
		sm:     manager,
		ss:     sync,
		ev:     evaluator.NewEvaluator(stores.splits, stores.segments, engine.NewEngine(logger), logger),
		is:     stores.impressions,
		iq:     impc.manager,
		cfg:    *c,
	}, nil
}

// Treatment implements Interface
func (i *Impl) Treatment(md *types.ClientMetadata, key string, bucketingKey string, feature string, attributes map[string]interface{}) (string, error) {
	res := i.ev.EvaluateFeature(key, &bucketingKey, feature, attributes)
	if res == nil {
		return "", fmt.Errorf("nil result")
	}

	err := i.handleImpression(key, bucketingKey, feature, res, *md)
	if err != nil {
		i.logger.Error("error handling impression: ", err)
	}

	return res.Treatment, nil
}

func (i *Impl) handleImpression(key string, bk string, f string, r *evaluator.Result, cm types.ClientMetadata) error {
	var label string
	if i.cfg.LabelsEnabled {
		label = r.Label
	}

	imp := dtos.Impression{
		FeatureName:  f,
		BucketingKey: bk,
		ChangeNumber: r.SplitChangeNumber,
		KeyName:      key,
		Label:        label,
		Treatment:    r.Treatment,
		Time:         timeMillis(),
	}

	shouldStore := i.iq.ProcessSingle(&imp)
	if shouldStore {
		_, err := i.is.Push(cm, imp)
		return err
	}

	return nil
}

func timeMillis() int64 {
	return time.Now().UTC().UnixMilli()
}

var _ Interface = (*Impl)(nil)
