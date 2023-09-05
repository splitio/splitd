package conf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfHelpers(t *testing.T) {
	var x int
	SetIfNotEmpty(&x, ref(0))
	assert.Equal(t, 0, x)
	SetIfNotNil(&x, nil)
	assert.Equal(t, 0, x)

	SetIfNotEmpty(&x, ref(5))
	assert.Equal(t, 5, x)
	SetIfNotNil(&x, ref(25))
	assert.Equal(t, 25, x)

    x = 0
	MapIfNotEmpty(&x, ref(0), func(z int) int { return z + 1 })
	assert.Equal(t, 0, x)
    MapIfNotNil(&x, nil, func(z int) int { return z + 1 })
	assert.Equal(t, 0, x)

	MapIfNotEmpty(&x, ref(1), func(z int) int { return z + 1 })
	assert.Equal(t, 2, x)
	MapIfNotEmpty(&x, ref(2), func(z int) int { return z + 1 })
	assert.Equal(t, 3, x)
}

func ref[T any](t T) *T {
	return &t
}
