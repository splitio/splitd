package errors

import (
	"errors"
	"sync"
)

type ConcurrentErrorCollector struct {
	errors []error
	mutex  sync.Mutex
}

func (c *ConcurrentErrorCollector) Append(err error) {
	c.mutex.Lock()
	c.errors = append(c.errors, err)
	c.mutex.Unlock()
}

func (c *ConcurrentErrorCollector) Join() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return errors.Join(c.errors...)
}
