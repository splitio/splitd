package types

import "github.com/splitio/go-split-commons/v4/dtos"

type ClientInterface interface {
	Treatment(key string, bucketingKey string, feature string, attrs map[string]interface{}) (*Result, error)
	Treatments(key string, bucketingKey string, features []string, attrs map[string]interface{}) (Results, error)
    Track(key string, trafficType string, eventType string, value *float64, properties map[string]interface{}) error
	Shutdown() error
}

type Result struct {
	Treatment  string
	Impression *dtos.Impression
}

type Results = map[string]Result
