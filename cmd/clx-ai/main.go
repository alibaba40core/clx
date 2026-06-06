package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/pipeline"
)

const workerEnvKey = "CLX_WORKER"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if os.Getenv(workerEnvKey) != "1" {
		fmt.Fprintf(stderr, "clx-ai is an internal component; invoke via clx\n")
		return 2
	}

	fs := flag.NewFlagSet("clx-ai", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "path to config.yaml (default: ~/.clx/config.yaml)")
	explain := fs.Bool("explain", false, "show intent and translation without executing")
	dryRun := fs.Bool("dry-run", false, "preview command without executing")
	yes := fs.Bool("y", false, "skip confirmation and execute")
	yesLong := fs.Bool("yes", false, "skip confirmation and execute")
	providerFlag := fs.String("provider", "", "override AI provider (none, ollama, openai, gemini)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(stderr, "clx-ai: missing command input")
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, cfg, closer, err := initWorker(ctx, *configPath, stderr)
	if err != nil {
		return 1
	}
	defer closer.Close()

	if *providerFlag != "" {
		cfg.Provider = *providerFlag
		cfg.Providers.Primary = *providerFlag
		cfg.Providers.Fallback = ""
		if err := config.Validate(cfg); err != nil {
			fmt.Fprintf(stderr, "config: %v\n", err)
			return 2
		}
	}

	rawInput := strings.Join(fs.Args(), " ")
	skipConfirm := *yes || *yesLong
	code, err := pipeline.Run(ctx, cfg, rawInput, pipeline.Options{
		Explain:  *explain,
		DryRun:   *dryRun,
		Yes:      skipConfirm,
		Logger:   logger,
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   stderr,
	})
	if err != nil && code == 0 {
		return 1
	}
	return code
}

func initWorker(ctx context.Context, configPath string, stderr io.Writer) (*slog.Logger, config.Config, io.Closer, error) {
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

type noopCloser struct{}

func (noopCloser) Close() error { return nil }
