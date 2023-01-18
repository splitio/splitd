package lfqueue

import (
	"sync/atomic"
	"unsafe"
)

type Interface[T any] interface {
	Push(T)
	Pop (T, bool)
}

type Impl[T any] struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	dummy lfqNode[T]
}

// NewLockfreeQueue is the only way to get a new, ready-to-use LockfreeQueue.
func NewLockfreeQueue[T any]() *Impl[T] {
	var lfq Impl[T]
	lfq.head = unsafe.Pointer(&lfq.dummy)
	lfq.tail = lfq.head
	return &lfq
}

func (lfq *Impl[T]) Pop() (T, bool) {
	for {
		h := atomic.LoadPointer(&lfq.head)
		rh := (*lfqNode[T])(h)
		n := (*lfqNode[T])(atomic.LoadPointer(&rh.next))
		if n != nil {
			if atomic.CompareAndSwapPointer(&lfq.head, h, rh.next) {
				return n.val, true
			} else {
				continue
			}
		} else {
			var v T
			return v, false
		}
	}
}

func (lfq *Impl[T]) Push(val T) {
	node := unsafe.Pointer(&lfqNode[T]{val: val})
	for {
		rt := (*lfqNode[T])(atomic.LoadPointer(&lfq.tail))
		if atomic.CompareAndSwapPointer(&rt.next, nil, node) {
			atomic.StorePointer(&lfq.tail, node)
			// If dead loop occurs, use CompareAndSwapPointer instead of StorePointer
			// atomic.CompareAndSwapPointer(&lfq.tail, t, node)
			return
		} else {
			continue
		}
	}
}

type lfqNode[T any] struct {
	val  T
	next unsafe.Pointer
}
