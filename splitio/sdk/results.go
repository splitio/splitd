package sdk

import "github.com/splitio/go-split-commons/v4/dtos"

type Result struct {
    Treatment string
    Impression *dtos.Impression
    Config *string
}
