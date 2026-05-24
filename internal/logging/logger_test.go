package logging

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestNewCreatesLogFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := config.LoggingConfig{Enabled: true, Level: "info"}

	logger, closer, err := New(context.Background(), cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	logger.Info("test message")

	logPath := filepath.Join(dir, "clx.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file missing: %v", err)
	}
}

func TestNewDisabledNoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := config.LoggingConfig{Enabled: false}

	logger, closer, err := New(context.Background(), cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	logger.Info("should not write")
	if _, err := os.Stat(filepath.Join(dir, "clx.log")); !os.IsNotExist(err) {
		t.Fatal("expected no log file when disabled")
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()
	if _, err := ParseLevel("debug"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseLevel("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCloserTwice(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := config.LoggingConfig{Enabled: true, Level: "info"}
	_, closer, err := New(context.Background(), cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestRotateIfNeeded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "clx.log")
	if err := os.WriteFile(logPath, make([]byte, maxLogBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := rotateIfNeeded(logPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatal("expected rotated backup")
	}
}
