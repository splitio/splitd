package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {

	e1 := RPCParseError{Code: PECWrongArgCount}
	assert.Equal(t, "wrong number of arguments for current opcode", e1.Error())

	e2 := RPCParseError{Code: PECOpCodeMismatch}
	assert.Equal(t, "opcode doesn't match the rpc whose arguments are being parsed", e2.Error())

	e3 := RPCParseError{Code: PECInvalidArgType, Data: 2}
	assert.Equal(t, "wrong argument type at index 2", e3.Error())

    e4 := RPCParseError{Code: RPCParseErrorCode(777)}
    assert.Equal(t, "unknown error", e4.Error())
}
