package logging

import (
	"fmt"
	"io"
	"os"

	"github.com/splitio/go-toolkit/v5/logging"
)

const (
	defaultMaxFiles    = 10
	defaultMaxFileSize = 1024 * 1024 // 1M
)

var defaultWriter = os.Stdout

func GetWriter(source *string, maxFiles *int, maxFileSize *int) (io.Writer, error) {
	if source == nil {
		return defaultWriter, nil
	}

	switch *source {
	case "stdout", "/dev/stdout":
		return os.Stdout, nil
	case "stderr", "/dev/stderr":
		return os.Stderr, nil
	default:
		// assume it's a regular file
		if maxFiles == nil && maxFileSize == nil {
			fileWriter, err := os.OpenFile(*source, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, fmt.Errorf("error creating log-output file: %w", err)
			}
			return fileWriter, nil
		}

		mf := valueOr(maxFiles, defaultMaxFiles)
		mfs := valueOr(maxFileSize, defaultMaxFileSize)
		return logging.NewFileRotate(&logging.FileRotateOptions{
			MaxBytes:    int64(mfs),
			BackupCount: mf,
			Path:        *source,
		})
	}
}

func valueOr[T any](t *T, fallback T) T {
	if t == nil {
		return fallback
	}
	return *t
}
