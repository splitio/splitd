package storage

import (
	"sync"
)

type MultiMetaQueues[T elemConstraint, U comparable, Q BackingQueue[T]] struct {
	m        sync.Map
	qpowSize int
	cFactory func() Q
}

func NewMultiMetaQueue[T elemConstraint, U comparable, Q BackingQueue[T]](cFactory func() Q) *MultiMetaQueues[T, U, Q] {
	return &MultiMetaQueues[T, U, Q]{
		m:        sync.Map{},
		cFactory: cFactory,
	}
}

func (m *MultiMetaQueues[T, U, Q]) Push(grouper U, items ...T) (int, error) {
	current, ok := m.m.Load(grouper)
	if !ok {
		q := m.cFactory()
		current, _ = m.m.LoadOrStore(grouper, q)
	}
	return current.(Q).Push(items...)
}

func (m *MultiMetaQueues[T, U, Q]) RangeAndClear(f func(U, Q)) error {
	m.m.Range(func(key, value any) bool {
		m.m.Delete(key)
		f(key.(U), value.(Q))
		return true
	})
	return nil
}

func (m *MultiMetaQueues[T, U, Q]) Range(f func(U, Q)) error {
	m.m.Range(func(key, value any) bool {
		f(key.(U), value.(Q))
		return true
	})
	return nil
}

type elemConstraint interface{}

type grupingConstraint comparable

type BackingQueue[T any] interface {
	Push(...T) (int, error)
}
