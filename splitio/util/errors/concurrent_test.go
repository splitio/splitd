package errors

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentErrors(t *testing.T) {

	// test setup:
	// errors are interfaces containing pointers, in order for them to be compared with errors.Is, it MUST be the same instance.
	// to do the test we create a slice with many vectors, use them and then compare against the original
	original := make([]error, 100)
	for idx := range original {
		original[idx] = errors.New(fmt.Sprintf("err_%d", idx))
	}

	var c ConcurrentErrorCollector

	var wg sync.WaitGroup
	wg.Add(100)
	for idx := 0; idx < 100; idx++ {
		go func(i int) {
			c.Append(original[i])
			wg.Done()
		}(idx)
	}

	wg.Wait()

	je := c.Join()
	for idx := 0; idx < 100; idx++ {
		assert.ErrorIs(t, je, original[idx])
	}

}
