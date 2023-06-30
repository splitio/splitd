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
	ErrReceivedSizeMismatch   = errors.New("recevied data size mismatch")
	ErrSenSentSizeMismatch    = errors.New("recevied data size mismatch")
)

type Interface interface {
	WriteFrame(writer io.Writer, raw []byte) (int, error)
	ReadFrame(reader io.Reader, readBuf []byte) (int, error)
}

type LengthPrefixImpl struct{}

func (l *LengthPrefixImpl) WriteFrame(writer io.Writer, raw []byte) (int, error) {
	size := len(raw)
	framedSize := size + 4
	framed := make([]byte, 0, framedSize)
	framed = encodeSize(size, framed)
	framed = append(framed, raw...)

	sent := 0
	for sent < framedSize {
        n, err := writer.Write(framed[sent:])
		if err != nil {
            return n, fmt.Errorf("error writing message: %w", err)
		}
		sent += n
	}

	if framedSize < 0 {
		return sent, ErrSenSentSizeMismatch
	}

	return sent - 4, nil

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

    size := decodeSize(sizeb[:])
	lr.N = int64(size)
	read := 0
	for read < int(size) {
		n, err = lr.Read(readBuf[read:])
		if err != nil {
			return n, fmt.Errorf("error reading message: %w", err)
		}
        read += n
	}

	if read != int(size) {
		return read, ErrReceivedSizeMismatch
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

