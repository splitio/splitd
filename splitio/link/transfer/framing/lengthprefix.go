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
	framed = binary.LittleEndian.AppendUint32(framed, uint32(size))
	framed = append(framed, raw...)

	n, err := writer.Write(framed)
	if err != nil {
		return n, fmt.Errorf("error writing message")
	}

	if n != framedSize {
		return n, ErrSenSentSizeMismatch
	}

	return n-4, nil

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
		return n, ErrReceivedSizeMismatch
	}

	return n, nil
}
