package storage

import (
	"math"

	"github.com/splitio/go-split-commons/v6/dtos"
	"github.com/splitio/splitd/splitio/sdk/types"
)

type ImpressionsStorage = MultiMetaQueues[dtos.Impression, types.ClientMetadata, *LockingQueue[dtos.Impression]]

type EventsStorage = MultiMetaQueues[dtos.EventDTO, types.ClientMetadata, *LockingQueue[dtos.EventDTO]]

func NewImpressionsQueue(approxSize int) (st *ImpressionsStorage, realSize int) {
	bits := getNearestSizePowerOf2(approxSize)
	return NewMultiMetaQueue[dtos.Impression, types.ClientMetadata](func() *LockingQueue[dtos.Impression] {
		return NewLKQueue[dtos.Impression](bits)
	}), int(math.Pow(2, float64(bits)))
}

func NewEventsQueue(approxSize int) (st *EventsStorage, realSize int) {
	bits := getNearestSizePowerOf2(approxSize)
	return NewMultiMetaQueue[dtos.EventDTO, types.ClientMetadata](func() *LockingQueue[dtos.EventDTO] {
		return NewLKQueue[dtos.EventDTO](bits)
	}), int(math.Pow(2, float64(bits)))
}

// to make the round-queue performant , we need to replace the modulo operation with an AND.
// For that approach to work, the size must be a power of 2. This function calculates the minimum power of 2
// that guarantees final_size >= requested_size
func getNearestSizePowerOf2(approxSize int) int {
	bits := int(math.Log2(float64(approxSize))) // floor(log2(approxSize))
	if math.Pow(2, float64(bits)) < float64(approxSize) {
		// if resulting size is lower than requested (because of float -> int conversion), add 1 bit
		bits++
	}
	return bits
}
