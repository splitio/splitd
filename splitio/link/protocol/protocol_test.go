package protocol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProtocolVersion(t *testing.T) {
	assert.Equal(t, "v1", Version(V1).String())
	assert.Equal(t, "invalid-version", Version(5).String())
}
