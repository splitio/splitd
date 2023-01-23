package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type K struct {
	V int
}

func TestLockingQueueBasic(t *testing.T) {

	st := NewLKQueue[K](2)

	n, err := st.Push(K{1}, K{2}, K{3})
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, st.Len())

	n, err = st.Push(K{4})
	assert.Equal(t, ErrQueueFull, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 3, st.Len())

	buf := make([]K, 0, 10)
	n, err = st.Pop(3, &buf)
	assert.Nil(t, err)
	assert.Equal(t, 3, n )
	assert.Equal(t, []K{{1}, {2}, {3}}, buf[:3])
	assert.Equal(t, 0, st.Len())

	n, err = st.Pop(1, &buf)
	assert.Equal(t, ErrQueueEmpty, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, st.Len())

	n, err = st.Push(K{4},K{5},K{6})
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, st.Len())
	
	n, err = st.Pop(3, &buf)
	assert.Nil(t, err)
	assert.Equal(t, 3, n )
	assert.Equal(t, []K{{4}, {5}, {6}}, buf[3:6])
	assert.Equal(t, 0, st.Len())

	n, err = st.Pop(1, &buf)
	assert.Equal(t, ErrQueueEmpty, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, st.Len())
}
