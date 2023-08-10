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
	"github.com/splitio/go-toolkit/v5/common"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
)

const (
	impressionsFullNotif = "IMPRESSIONS_FULL"
)

type Attributes = map[string]interface{}

type Interface interface {
	Treatment(cfg *types.ClientConfig, Key string, BucketingKey *string, Feature string, attributes map[string]interface{}) (*Result, error)
	Treatments(cfg *types.ClientConfig, Key string, BucketingKey *string, Features []string, attributes map[string]interface{}) (map[string]Result, error)
}

type Impl struct {
	logger        logging.LoggerInterface
	ev            evaluator.Interface
	sm            synchronizer.Manager
	ss            synchronizer.Synchronizer
	is            *storage.ImpressionsStorage
	iq            provisional.ImpressionManager
	cfg           sdkConf.Config
	status        chan int
	queueFullChan chan string
}

func New(logger logging.LoggerInterface, apikey string, c *conf.Config) (*Impl, error) {

	if warnings := c.Normalize(); len(warnings) > 0 {
		for _, w := range warnings {
			logger.Warning(w)
		}
	}

	md := dtos.Metadata{SDKVersion: fmt.Sprintf("splitd-%s", splitio.Version)}
	advCfg := c.ToAdvancedConfig()

	stores := setupStorages(c)
	impc, err := setupImpressionsComponents(&c.Impressions, stores.telemetry)
	if err != nil {
		return nil, fmt.Errorf("error setting up impressions components")
	}

	hc := &application.Dummy{}

	queueFullChan := make(chan string, 1) // Only one item so that we don't queue N flushes (which makes no sense) if we're getting hit too hard
	splitApi := api.NewSplitAPI(apikey, *advCfg, logger, md)
	workers := setupWorkers(logger, splitApi, stores, hc, &c.Impressions)
	tasks := setupTasks(c, stores, logger, workers, impc, md, splitApi)
	sync := synchronizer.NewSynchronizer(*advCfg, *tasks, *workers, logger, queueFullChan, nil)

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
		logger:        logger,
		sm:            manager,
		ss:            sync,
		ev:            evaluator.NewEvaluator(stores.splits, stores.segments, engine.NewEngine(logger), logger),
		is:            stores.impressions,
		iq:            impc.manager,
		cfg:           *c,
		queueFullChan: queueFullChan,
	}, nil
}

// Treatment implements Interface
func (i *Impl) Treatment(cfg *types.ClientConfig, key string, bk *string, feature string, attributes Attributes) (*Result, error) {
	res := i.ev.EvaluateFeature(key, bk, feature, attributes)
	if res == nil {
		return nil, fmt.Errorf("nil result")
	}

	imp := i.handleImpression(key, bk, feature, res, cfg.Metadata)
	return &Result{
		Treatment:  res.Treatment,
		Impression: imp,
		Config:     res.Config,
	}, nil
}

// Treatment implements Interface
func (i *Impl) Treatments(cfg *types.ClientConfig, key string, bk *string, features []string, attributes Attributes) (map[string]Result, error) {

	res := i.ev.EvaluateFeatures(key, bk, features, attributes)
	toRet := make(map[string]Result, len(res.Evaluations))
	for _, feature := range features {

		curr, ok := res.Evaluations[feature]
		if !ok {
			toRet[feature] = Result{Treatment: "control"}
			continue
		}

		var eres Result
		eres.Treatment = curr.Treatment
		eres.Impression = i.handleImpression(key, bk, feature, &curr, cfg.Metadata)
		eres.Config = curr.Config
		toRet[feature] = eres
	}

	return toRet, nil
}

func (i *Impl) handleImpression(key string, bk *string, f string, r *evaluator.Result, cm types.ClientMetadata) *dtos.Impression {
	var label string
	if i.cfg.LabelsEnabled {
		label = r.Label
	}

	imp := &dtos.Impression{
		FeatureName:  f,
		BucketingKey: common.StringFromRef(bk),
		ChangeNumber: r.SplitChangeNumber,
		KeyName:      key,
		Label:        label,
		Treatment:    r.Treatment,
		Time:         timeMillis(),
	}

	shouldStore := i.iq.ProcessSingle(imp)
	if shouldStore {
		_, err := i.is.Push(cm, *imp)
		if err != nil {
			if err == storage.ErrQueueFull {
				select {
				case i.queueFullChan <- impressionsFullNotif:
				default:
					i.logger.Warning("impressions queue has filled up and is currently performing a flush. Current impression will bedropped")
				}
			} else {
				i.logger.Error("error handling impression: ", err)
			}
		}
	}

	return imp
}

func timeMillis() int64 {
	return time.Now().UTC().UnixMilli()
}

var _ Interface = (*Impl)(nil)
