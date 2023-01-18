package framing

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInsufficientHeaderData = errors.New("not enough bytes received when decoding packet size")
	ErrSizeDecode             = errors.New("error decoding size prefix from message")
	ErrSizeMismatch           = errors.New("recevied data size mismatch")
)

type Interface interface {
	Frame(raw []byte) []byte
	ReadFrame(reader io.Reader, readBuf []byte) (int, error)
}

type LengthPrefixImpl struct{}

func (l *LengthPrefixImpl) Frame(raw []byte) []byte {
	size := len(raw)
	dst := make([]byte, 0, size+4)
	dst = binary.LittleEndian.AppendUint32(dst, uint32(size))
	dst = append(dst, raw...)
	return dst
}

func (l *LengthPrefixImpl) ReadFrame(reader io.Reader, readBuf []byte) (int, error) {
	lr := io.LimitedReader{R: reader, N: 4}
	var sizeb [4]byte
	n, err := lr.Read(sizeb[:])
	if err != nil {
		return 0, fmt.Errorf("error reading size: %w", err)
	}

	if n != 4 {
		return 0, ErrInsufficientHeaderData
	}

	size := binary.LittleEndian.Uint32(sizeb[:])

	lr.N = int64(size)
	n, err = lr.Read(readBuf[:])
	if err != nil {
		return n, fmt.Errorf("error reading message: %w", err)
	}

	if n != int(size) {
		return n, ErrSizeMismatch
	}

	return n, nil
}
