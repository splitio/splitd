package sdk

import (
	"errors"
	"fmt"
	"time"

	"github.com/splitio/splitd/splitio/sdk/conf"
	"github.com/splitio/splitd/splitio/sdk/storage"
	"github.com/splitio/splitd/splitio/sdk/types"

	"github.com/splitio/go-split-commons/v5/dtos"
	"github.com/splitio/go-split-commons/v5/engine"
	"github.com/splitio/go-split-commons/v5/engine/evaluator"
	"github.com/splitio/go-split-commons/v5/flagsets"
	"github.com/splitio/go-split-commons/v5/healthcheck/application"
	"github.com/splitio/go-split-commons/v5/provisional"
	"github.com/splitio/go-split-commons/v5/service/api"
	commonStorage "github.com/splitio/go-split-commons/v5/storage"
	"github.com/splitio/go-split-commons/v5/synchronizer"
	"github.com/splitio/go-toolkit/v5/common"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/splitd/splitio"
)

const (
	impressionsFullNotif = "IMPRESSIONS_FULL"
	eventsFullNotif      = "EVENTS_FULL"
)

var (
	ErrEventsQueueFull = errors.New("events queue full")
	ErrSplitNotFound   = errors.New("split not found")
)

type Attributes = map[string]interface{}

type Interface interface {
	Treatment(cfg *types.ClientConfig, key string, bucketingKey *string, feature string, attributes map[string]interface{}) (*EvaluationResult, error)
	Treatments(cfg *types.ClientConfig, key string, bucketingKey *string, features []string, attributes map[string]interface{}) (map[string]EvaluationResult, error)
	TreatmentsByFlagSet(cfg *types.ClientConfig, key string, bucketingKey *string, flagSet string, attributes map[string]interface{}) (map[string]EvaluationResult, error)
	TreatmentsByFlagSets(cfg *types.ClientConfig, key string, bucketingKey *string, flagSets []string, attributes map[string]interface{}) (map[string]EvaluationResult, error)
	Track(cfg *types.ClientConfig, key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error
	SplitNames() ([]string, error)
	Splits() ([]SplitView, error)
	Split(name string) (*SplitView, error)
	Shutdown() error
}

type Impl struct {
	logger        logging.LoggerInterface
	ev            evaluator.Interface
	sm            synchronizer.Manager
	ss            synchronizer.Synchronizer
	is            *storage.ImpressionsStorage
	es            *storage.EventsStorage
	iq            provisional.ImpressionManager
	splitStorage  commonStorage.SplitStorage
	cfg           conf.Config
	queueFullChan chan string
	validator     Validator
}

func New(logger logging.LoggerInterface, apikey string, c *conf.Config) (*Impl, error) {

	if warnings := c.Normalize(); len(warnings) > 0 {
		for _, w := range warnings {
			logger.Warning(w)
		}
	}

	md := dtos.Metadata{SDKVersion: fmt.Sprintf("splitd-%s", splitio.Version)}
	advCfg := c.ToAdvancedConfig()

	flagSets, errs := flagsets.SanitizeMany(advCfg.FlagSetsFilter)
	if len(errs) != 0 {
		for _, err := range errs {
			if errType, ok := err.(dtos.FlagSetValidatonError); ok {
				logger.Warning(errType.Message)
			}
		}
	}
	flagSetsFilter := flagsets.NewFlagSetFilter(flagSets)

	stores := setupStorages(c, flagSetsFilter)
	impc, err := setupImpressionsComponents(&c.Impressions, stores.telemetry)
	if err != nil {
		return nil, fmt.Errorf("error setting up impressions components")
	}

	hc := &application.Dummy{}

	queueFullChan := make(chan string, 2)
	splitApi := api.NewSplitAPI(apikey, *advCfg, logger, md)
	workers := setupWorkers(logger, splitApi, stores, hc, c, flagSetsFilter)
	tasks := setupTasks(c, stores, logger, workers, impc, md, splitApi)
	sync := synchronizer.NewSynchronizer(*advCfg, *tasks, *workers, logger, queueFullChan)

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
		es:            stores.events,
		iq:            impc.manager,
		splitStorage:  stores.splits,
		cfg:           *c,
		queueFullChan: queueFullChan,
		validator:     Validator{logger: logger, splits: stores.splits},
	}, nil
}

// Treatment implements Interface
func (i *Impl) Treatment(cfg *types.ClientConfig, key string, bk *string, feature string, attributes Attributes) (*EvaluationResult, error) {
	res := i.ev.EvaluateFeature(key, bk, feature, attributes)
	if res == nil {
		return nil, fmt.Errorf("nil result")
	}

	imp := i.handleImpression(key, bk, feature, res, cfg.Metadata)
	return &EvaluationResult{
		Treatment:  res.Treatment,
		Impression: imp,
		Config:     res.Config,
	}, nil
}

// Treatment implements Interface
func (i *Impl) Treatments(cfg *types.ClientConfig, key string, bk *string, features []string, attributes Attributes) (map[string]EvaluationResult, error) {

	res := i.ev.EvaluateFeatures(key, bk, features, attributes)
	toRet := make(map[string]EvaluationResult, len(res.Evaluations))
	for _, feature := range features {

		curr, ok := res.Evaluations[feature]
		if !ok {
			toRet[feature] = EvaluationResult{Treatment: "control"}
			continue
		}

		var eres EvaluationResult
		eres.Treatment = curr.Treatment
		eres.Impression = i.handleImpression(key, bk, feature, &curr, cfg.Metadata)
		eres.Config = curr.Config
		toRet[feature] = eres
	}

	return toRet, nil
}

// TreatmentsByFlagSet implements Interface
func (i *Impl) TreatmentsByFlagSet(cfg *types.ClientConfig, key string, bk *string, flagSet string, attributes Attributes) (map[string]EvaluationResult, error) {

	res := i.ev.EvaluateFeatureByFlagSets(key, bk, []string{flagSet}, attributes)
	toRet := make(map[string]EvaluationResult, len(res.Evaluations))
	for feature, curr := range res.Evaluations {
		var eres EvaluationResult
		eres.Treatment = curr.Treatment
		eres.Impression = i.handleImpression(key, bk, feature, &curr, cfg.Metadata)
		eres.Config = curr.Config
		toRet[feature] = eres
	}

	return toRet, nil
}

// TreatmentsByFlagSets implements Interface
func (i *Impl) TreatmentsByFlagSets(cfg *types.ClientConfig, key string, bk *string, flagSets []string, attributes Attributes) (map[string]EvaluationResult, error) {

	res := i.ev.EvaluateFeatureByFlagSets(key, bk, flagSets, attributes)
	toRet := make(map[string]EvaluationResult, len(res.Evaluations))
	for feature, curr := range res.Evaluations {
		var eres EvaluationResult
		eres.Treatment = curr.Treatment
		eres.Impression = i.handleImpression(key, bk, feature, &curr, cfg.Metadata)
		eres.Config = curr.Config
		toRet[feature] = eres
	}

	return toRet, nil
}

func (i *Impl) Track(cfg *types.ClientConfig, key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error {

	// TODO(mredolatti): validate traffic type & truncate properties if needed
	trafficType, err := i.validator.validateTrafficType(trafficType)
	if err != nil {
		return err
	}

	properties, _, err = i.validator.validateTrackProperties(properties)
	if err != nil {
		return err
	}

	event := &dtos.EventDTO{
		Key:             key,
		TrafficTypeName: trafficType,
		EventTypeID:     eventType,
		Value:           value,
		Timestamp:       timeMillis(),
		Properties:      properties,
	}

	_, err = i.es.Push(cfg.Metadata, *event)
	if err != nil {
		if err == storage.ErrQueueFull {
			select {
			case i.queueFullChan <- eventsFullNotif:
			default:
				i.logger.Warning("events queue has filled up and is currently performing a flush. Current event will be dropped")
			}
			return ErrEventsQueueFull
		}
		i.logger.Error("error handling event: ", err)
		return err
	}
	return nil
}

func (i *Impl) Split(name string) (*SplitView, error) {
	split := i.splitStorage.Split(name)
	if split == nil {
		return nil, ErrSplitNotFound
	}

	return splitToView(split), nil
}

func (i *Impl) SplitNames() ([]string, error) {
	return i.splitStorage.SplitNames(), nil
}

func (i *Impl) Splits() ([]SplitView, error) {
	splits := i.splitStorage.All()
	asViews := make([]SplitView, 0, len(splits))
	for idx := range splits {
		asViews = append(asViews, *splitToView(&splits[idx]))
	}
	return asViews, nil
}

func (i *Impl) Shutdown() error {
	i.sm.Stop()
	return nil
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

func splitToView(s *dtos.SplitDTO) *SplitView {

	var treatments []string
	for _, c := range s.Conditions {
		for _, p := range c.Partitions {
			treatments = append(treatments, p.Treatment)
		}
	}

	return &SplitView{
		Name:             s.Name,
		TrafficType:      s.TrafficTypeName,
		Killed:           s.Killed,
		ChangeNumber:     s.ChangeNumber,
		Configs:          s.Configurations,
		Treatments:       treatments,
		DefaultTreatment: s.DefaultTreatment,
		Sets:             s.Sets,
	}
}

func timeMillis() int64 {
	return time.Now().UTC().UnixMilli()
}

var _ Interface = (*Impl)(nil)
