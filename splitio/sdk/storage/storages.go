package storage

import (
	"math"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/splitd/splitio/sdk/types"
)

type ImpressionsStorage = MultiMetaQueues[dtos.Impression, types.ClientMetadata, *LockingQueue[dtos.Impression]]

type EventsStorage = MultiMetaQueues[dtos.EventDTO, types.ClientMetadata, *LockingQueue[dtos.EventDTO]]

func NewImpressionsQueue(approxSize int) (st *ImpressionsStorage, realSize int) {
	bits := int(math.Log2(float64(approxSize))) + 1
	return NewMultiMetaQueue[dtos.Impression, types.ClientMetadata](func() *LockingQueue[dtos.Impression] {
		return NewLKQueue[dtos.Impression](bits)
	}), int(math.Pow(2, float64(bits)))
}

func NewEventsQueue(approxSize int) (st *EventsStorage, realSize int) {
	bits := int(math.Log2(float64(approxSize))) + 1
	return NewMultiMetaQueue[dtos.EventDTO, types.ClientMetadata](func() *LockingQueue[dtos.EventDTO] {
		return NewLKQueue[dtos.EventDTO](bits)
	}), int(math.Pow(2, float64(bits)))
}

