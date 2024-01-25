package sdk

import "github.com/splitio/go-split-commons/v5/dtos"

type EvaluationResult struct {
	Treatment  string
	Impression *dtos.Impression
	Config     *string
}

type SplitView struct {
	Name             string
	TrafficType      string
	Killed           bool
	Treatments       []string
	ChangeNumber     int64
	Configs          map[string]string
	DefaultTreatment string
	Sets             []string
}
