package serializer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnType(t *testing.T) {
	assert.Equal(t, "msgpack", MsgPack.String())
	assert.Equal(t, "invalid-serialization", Mechanism(123).String())
}

func TestSetup(t *testing.T) {
	_, err := Setup(MsgPack)
	assert.Nil(t, err)

	_, err = Setup(Mechanism(123))
	assert.ErrorContains(t, err, "unknown serialization mechanism")
}
