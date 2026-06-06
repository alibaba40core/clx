package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/logging"
)

// initMinimal loads config only (no bootstrap, discard logger). Used on the hot
// pipeline path to avoid FS setup and log file I/O per invocation.
func initMinimal(ctx context.Context, configPath string, stderr io.Writer) (*slog.Logger, config.Config, io.Closer, error) {
	path := configPath
	var err error
	if path == "" {
		path, err = config.ConfigPath()
		if err != nil {
			fmt.Fprintf(stderr, "config path: %v\n", err)
			return nil, config.Config{}, nil, err
		}
	}

	cfg, err := config.Load(ctx, path)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return nil, config.Config{}, nil, err
	}

	logger := slog.New(slog.DiscardHandler)
	return logger, cfg, noopCloser{}, nil
}

// initFull runs bootstrap, loads config, and opens the log file when enabled.
// Used by doctor, init, and other subcommands that write under ~/.clx/.
func initFull(ctx context.Context, configPath string, stderr io.Writer) (*slog.Logger, config.Config, io.Closer, error) {
	result, err := config.Bootstrap(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return nil, config.Config{}, nil, err
	}
	if result.WroteConfig || result.WrotePolicy {
		fmt.Fprintln(stderr, "CLX: first-run setup complete (~/.clx/)")
	}

	path := configPath
	if path == "" {
		path, err = config.ConfigPath()
		if err != nil {
			fmt.Fprintf(stderr, "config path: %v\n", err)
			return nil, config.Config{}, nil, err
		}
	}

	cfg, err := config.Load(ctx, path)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return nil, config.Config{}, nil, err
	}

	logsDir, err := config.LogsDir()
	if err != nil {
		fmt.Fprintf(stderr, "logs dir: %v\n", err)
		return nil, config.Config{}, nil, err
	}

	logger, closer, err := logging.New(ctx, cfg.Logging, logsDir)
	if err != nil {
		fmt.Fprintf(stderr, "logger: %v\n", err)
		return nil, config.Config{}, nil, err
	}
	return logger, cfg, closer, nil
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }
