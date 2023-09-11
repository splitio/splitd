package lang

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfHelpers(t *testing.T) {
	var x int
	SetIfNotEmpty(&x, Ref(0))
	assert.Equal(t, 0, x)
	SetIfNotNil(&x, nil)
	assert.Equal(t, 0, x)

	SetIfNotEmpty(&x, Ref(5))
	assert.Equal(t, 5, x)
	SetIfNotNil(&x, Ref(25))
	assert.Equal(t, 25, x)

	x = 0
	MapIfNotEmpty(&x, Ref(0), func(z int) int { return z + 1 })
	assert.Equal(t, 0, x)
	MapIfNotNil(&x, nil, func(z int) int { return z + 1 })
	assert.Equal(t, 0, x)

	MapIfNotEmpty(&x, Ref(1), func(z int) int { return z + 1 })
	assert.Equal(t, 2, x)
	MapIfNotEmpty(&x, Ref(2), func(z int) int { return z + 1 })
	assert.Equal(t, 3, x)
}
