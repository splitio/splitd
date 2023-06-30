package storage

import (
	"errors"
	"math"
	"sync"
)

var (
	ErrQueueFull  = errors.New("queue full")
	ErrQueueEmpty = errors.New("queue empty")
)

type LockingQueue[T any] struct {
	data  []T
	pow   int
	mask  int
	head  int
	tail  int
	mutex sync.Mutex
}

func NewLKQueue[T any](pow int) *LockingQueue[T] {
	bufSize := int(math.Pow(2, float64(pow)))
	return &LockingQueue[T]{
		data: make([]T, bufSize),
		pow:  pow,
		mask: bufSize - 1,
	}
}

func (q *LockingQueue[T]) Push(ts ...T) (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for added, t := range ts {
		if !q.queue(t) {
			return added, ErrQueueFull
		}
	}
	return len(ts), nil
}

func (q *LockingQueue[T]) Pop(n int, buf *[]T) (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	nc := n
	for nc > 0 {
		elem, ok := q.dequeue()
		if !ok {
			return n - nc, ErrQueueEmpty
		}
		*buf = append(*buf, elem)
		nc--
	}
	return n, nil
}

func (q *LockingQueue[T]) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return (q.head - q.tail) & q.mask
}

func (q *LockingQueue[T]) queue(t T) bool {
	nextpos := ((q.head + 1) & q.mask)
	if nextpos == q.tail {
		return false // queue full
	}

	q.data[q.head] = t
	q.head = nextpos
	return true
}

func (q *LockingQueue[T]) dequeue() (T, bool) {
	if q.tail == q.head {
		var t T
		return t, false
	}

	tmp := q.data[q.tail]
	q.tail = ((q.tail + 1) & q.mask)
	return tmp, true
}
