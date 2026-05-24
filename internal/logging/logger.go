package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alibaba40core/clx/internal/config"
)

const maxLogBytes = 5 * 1024 * 1024

// New creates a slog.Logger writing JSON lines to dir/clx.log.
// When cfg.Enabled is false, returns a discard logger and a no-op closer.
func New(ctx context.Context, cfg config.LoggingConfig, dir string) (*slog.Logger, io.Closer, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if !cfg.Enabled {
		return slog.New(slog.DiscardHandler), noopCloser{}, nil
	}

	level, err := ParseLevel(cfg.Level)
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, nil, err
	}

	logPath := filepath.Join(dir, "clx.log")
	if err := rotateIfNeeded(logPath); err != nil {
		return nil, nil, err
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, err
	}

	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	return logger, &fileCloser{f: f}, nil
}

type fileCloser struct {
	f *os.File
}

func (c *fileCloser) Close() error {
	if c.f == nil {
		return nil
	}
	err := c.f.Close()
	c.f = nil
	return err
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }

func rotateIfNeeded(logPath string) error {
	info, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size() <= maxLogBytes {
		return nil
	}
	backup := logPath + ".1"
	_ = os.Remove(backup)
	return os.Rename(logPath, backup)
}
