package storage

import (
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/splitd/splitio/sdk/types"

	"github.com/stretchr/testify/assert"
)

func TestMultiStorageBasic(t *testing.T) {

	mq := NewMultiMetaQueue[dtos.EventDTO, types.ClientMetadata](func() *LockingQueue[dtos.EventDTO] { return NewLKQueue[dtos.EventDTO](4) })

	n, err := mq.Push(types.ClientMetadata{SdkVersion: "go-1.2.3"}, dtos.EventDTO{Key: "k1"})
	assert.Equal(t, 1, n)
	assert.Nil(t, err)

	n, err = mq.Push(types.ClientMetadata{SdkVersion: "go-1.2.3"}, dtos.EventDTO{Key: "k2"}, dtos.EventDTO{Key: "k3"}, dtos.EventDTO{Key: "k4"})
	assert.Equal(t, 3, n)
	assert.Nil(t, err)

	n, err = mq.Push(types.ClientMetadata{SdkVersion: "php-1.2.3"}, dtos.EventDTO{Key: "k2"}, dtos.EventDTO{Key: "k3"}, dtos.EventDTO{Key: "k4"})
	assert.Equal(t, 3, n)
	assert.Nil(t, err)

	mq.RangeAndClear(func(cm types.ClientMetadata, q *LockingQueue[dtos.EventDTO]) {
		switch cm.SdkVersion {
		case "go-1.2.3":
			assert.Equal(t, 4, q.Len())
			var buf []dtos.EventDTO
			n, err := q.Pop(5, &buf)
			assert.Equal(t, ErrQueueEmpty, err)
			assert.Equal(t, 4, n)
		case "php-1.2.3":
			assert.Equal(t, 3, q.Len())
			var buf []dtos.EventDTO
			n, err := q.Pop(5, &buf)
			assert.Equal(t, ErrQueueEmpty, err)
			assert.Equal(t, 3, n)
		default:
			assert.Fail(t, "unexpected metadata: "+cm.SdkVersion)
		}
	})

	mq.RangeAndClear(func(cm types.ClientMetadata, q *LockingQueue[dtos.EventDTO]) {
		assert.Fail(t, "should not execute")
	})
}
