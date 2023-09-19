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
)

type Interface interface {
	WriteFrame(raw []byte) (int, error)
	ReadFrame(readBuf []byte) (int, error)
}

type LengthPrefixImpl struct {
	lr io.LimitedReader
	w  io.Writer
}

func NewLengthPrefix(rw io.ReadWriter) *LengthPrefixImpl {
	return &LengthPrefixImpl{
		lr: io.LimitedReader{R: rw, N: 4},
		w:  rw,
	}
}

func (l *LengthPrefixImpl) WriteFrame(raw []byte) (int, error) {
	size := len(raw)
	framedSize := size + 4
	framed := make([]byte, 0, framedSize)
	framed = encodeSize(size, framed)
	framed = append(framed, raw...)

	sent := 0
	for sent < framedSize {
		n, err := l.w.Write(framed[sent:])
		sent += n
		if err != nil {
			return sent, err
		}
	}

	return sent - 4, nil

}

func (l *LengthPrefixImpl) ReadFrame(readBuf []byte) (int, error) {
	var sizeb [4]byte
	l.lr.N = 4
	n, err := l.lr.Read(sizeb[:])
	if err != nil {
		return 0, fmt.Errorf("error reading size: %w", err)
	}

	if n != 4 {
		return 0, ErrInsufficientHeaderData
	}

	size := decodeSize(sizeb[:])
	if bufSize := len(readBuf); int(size) > bufSize {
		return 0, fmt.Errorf("read buffer is too small (%d bytes) to handle incoming message (%d bytes)", bufSize, size)
	}

	l.lr.N = int64(size)

	read := 0
	for read < int(size) {
		n, err = l.lr.Read(readBuf[read:])
		read += n
		if err != nil {
			return read, err
		}
	}

	return read, nil
}

func encodeSize(size int, target []byte) []byte {
	target = binary.LittleEndian.AppendUint32(target, uint32(size))
	return target
}
func decodeSize(size []byte) uint32 {
	return binary.LittleEndian.Uint32(size[:])
}

var _ Interface = (*LengthPrefixImpl)(nil)
