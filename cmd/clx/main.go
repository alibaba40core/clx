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

	"github.com/alibaba40core/clx/internal/cliversion"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/logging"
	"github.com/alibaba40core/clx/internal/pipeline"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "doctor" {
		return runDoctor(args[1:], stdout, stderr)
	}

	fs := flag.NewFlagSet("clx", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	showHelp := fs.Bool("help", false, "show help")
	configPath := fs.String("config", "", "path to config.yaml (default: ~/.clx/config.yaml)")
	explain := fs.Bool("explain", false, "show intent and translation without executing")
	dryRun := fs.Bool("dry-run", false, "preview command without executing")
	yes := fs.Bool("y", false, "skip confirmation and execute")
	yesLong := fs.Bool("yes", false, "skip confirmation and execute")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showHelp {
		printHelp(stdout)
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, cfg, closer, err := initCLX(ctx, *configPath, stderr)
	if err != nil {
		return 1
	}
	defer closer.Close()

	logger.Debug("clx invoked", "version", cliversion.Version)

	if *showVersion {
		fmt.Fprintln(stdout, cliversion.Line("clx"))
		return 0
	}

	if fs.NArg() == 0 {
		printHelp(stdout)
		return 0
	}

	rawInput := strings.Join(fs.Args(), " ")
	skipConfirm := *yes || *yesLong
	code, err := pipeline.Run(ctx, cfg, rawInput, pipeline.Options{
		Explain: *explain,
		DryRun:  *dryRun,
		Yes:     skipConfirm,
		Stdout:  stdout,
		Stderr:  stderr,
	})
	if err != nil && code == 0 {
		return 1
	}
	return code
}

func runDoctor(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	configPath := ""
	if len(args) > 0 && args[0] == "--config" && len(args) > 1 {
		configPath = args[1]
	}

	logger, _, closer, err := initCLX(ctx, configPath, stderr)
	if err != nil {
		return 1
	}
	defer closer.Close()

	logger.Info("clx doctor started")
	if err := environment.RunDoctor(ctx, stdout); err != nil {
		fmt.Fprintf(stderr, "doctor: %v\n", err)
		return 1
	}
	return 0
}

func initCLX(ctx context.Context, configPath string, stderr io.Writer) (*slog.Logger, config.Config, io.Closer, error) {
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

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "CLX — cross-platform command intelligence")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx [flags] <command...>")
	fmt.Fprintln(w, "  clx doctor")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  clx grep errors logs.txt")
	fmt.Fprintln(w, "  clx --explain pwd")
	fmt.Fprintln(w, "  clx --dry-run ls .")
	fmt.Fprintln(w, "  clx -y pwd")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  doctor          Detect environment and write ~/.clx/system_profile.json")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --explain       Show intent and translation without executing")
	fmt.Fprintln(w, "  --dry-run       Preview command without executing")
	fmt.Fprintln(w, "  -y, --yes       Skip confirmation and execute")
	fmt.Fprintln(w, "  --version       Print version and exit")
	fmt.Fprintln(w, "  --help          Show this help")
	fmt.Fprintln(w, "  --config path   Path to config.yaml")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Exit codes: 0 success, 1 error, 2 flag error")
}
