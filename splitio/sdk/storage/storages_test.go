package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNearestSizePowerOf2(t *testing.T) {
    assert.Equal(t, 1, getNearestSizePowerOf2(2))
    assert.Equal(t, 2, getNearestSizePowerOf2(3))
    assert.Equal(t, 2, getNearestSizePowerOf2(4))
    assert.Equal(t, 3, getNearestSizePowerOf2(5))
    assert.Equal(t, 3, getNearestSizePowerOf2(8))
    assert.Equal(t, 10, getNearestSizePowerOf2(1024))
    assert.Equal(t, 11, getNearestSizePowerOf2(1025))
}
