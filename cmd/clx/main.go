package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/alibaba40core/clx/internal/cliversion"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/logging"
)

const phaseMessage = "CLX command pipeline is not available yet (Phase 1.6). Try: clx --version"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("clx", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	showHelp := fs.Bool("help", false, "show help")
	configPath := fs.String("config", "", "path to config.yaml (default: ~/.clx/config.yaml)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showHelp {
		printHelp(stdout)
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	result, err := config.Bootstrap(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}
	if result.WroteConfig || result.WrotePolicy {
		fmt.Fprintln(stderr, "CLX: first-run setup complete (~/.clx/)")
	}

	path := *configPath
	if path == "" {
		path, err = config.ConfigPath()
		if err != nil {
			fmt.Fprintf(stderr, "config path: %v\n", err)
			return 1
		}
	}

	cfg, err := config.Load(ctx, path)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}

	logsDir, err := config.LogsDir()
	if err != nil {
		fmt.Fprintf(stderr, "logs dir: %v\n", err)
		return 1
	}

	logger, closer, err := logging.New(ctx, cfg.Logging, logsDir)
	if err != nil {
		fmt.Fprintf(stderr, "logger: %v\n", err)
		return 1
	}
	defer closer.Close()

	logger.Debug("clx invoked", "version", cliversion.Version)

	if *showVersion {
		fmt.Fprintln(stdout, cliversion.Line("clx"))
		return 0
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "%s\n", phaseMessage)
		return 0
	}

	fmt.Fprintln(stdout, phaseMessage)
	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "CLX — cross-platform command intelligence")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --version       Print version and exit")
	fmt.Fprintln(w, "  --help          Show this help")
	fmt.Fprintln(w, "  --config path   Path to config.yaml")
}
