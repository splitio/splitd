package types

import (
	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/splitd/splitio/sdk"
)

type ClientInterface interface {
	Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (*Result, error)
	Treatments(key string, bucketingKey string, features []string, attrs map[string]interface{}) (Results, error)
	TreatmentWithConfig(key string, bucketingKey string, feature string, attrs map[string]interface{}) (*Result, error)
	TreatmentsWithConfig(key string, bucketingKey string, features []string, attrs map[string]interface{}) (Results, error)
	Track(key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error
	SplitNames() ([]string, error)
	Split(name string) (*sdk.SplitView, error)
	Splits() ([]sdk.SplitView, error)
	Shutdown() error
}

type Result struct {
	Treatment  string
	Impression *dtos.Impression
	Config     *string
}

type Results = map[string]Result
