package v1

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrInternal           = errors.New("internal agent error")
	ErrIncorrectArguments = errors.New("opcode doesn't match arguments type")
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

var _ ErrorWithResponse = (*InvocationError)(nil)
