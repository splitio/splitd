package v1

import (
	"errors"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrInternal          = errors.New("internal agent error")
	ErrOpcodeArgMismatch = errors.New("opcode doesn't match arguments type")
	ErrIncorrectArgCount = errors.New("invalid argument count")
)

type InvocationErrorCode int

const (
	InvocationErrorInvalidArgs InvocationErrorCode = 0
)

type ErrorWithResponse interface {
	ToResponse() []byte
}

type InvocationError struct {
	code    InvocationErrorCode
	message string
}

func (e *InvocationError) Error() string {
	return e.message
}

func (e *InvocationError) ToResponse() []byte {
	serialized, _ := msgpack.Marshal(e)
	return serialized
}

// ---------------

type RPCParseErrorCode int

const (
	PECOpCodeMismatch = 1
	PECWrongArgCount  = 2
	PECInvalidArgType = 3
)

func (c RPCParseErrorCode) formatWithData(data int64) string {
	switch c {
	case PECOpCodeMismatch:
		return "opcode doesn't match the rpc whose arguments are being parsed"
	case PECWrongArgCount:
		return "wrong number of arguments for current opcode"
	case PECInvalidArgType:
		return "wrong argument type at index " + strconv.Itoa(int(data))
	default:
        return "unknown error"
	}
}

type RPCParseError struct {
	Code RPCParseErrorCode
	Data int64
}

// Error implements error
func (e RPCParseError) Error() string {
    return e.Code.formatWithData(e.Data)
}

var _ error = RPCParseError{}

// ----------------
var _ ErrorWithResponse = (*InvocationError)(nil)
