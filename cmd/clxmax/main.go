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

const phaseMessage = "clxmax reasoning mode (Version 2) is in development — see doc/phase-5.md. Try: clxmax --version"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("clxmax", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	showHelp := fs.Bool("help", false, "show help")
	configPath := fs.String("config", "", "path to config.yaml")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showHelp {
		fmt.Fprintln(stdout, "clxmax — advanced reasoning mode (Version 2 / Phase 5)")
		fmt.Fprintln(stdout, "  clxmax --version")
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if _, err := config.Bootstrap(ctx); err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}

	path := *configPath
	var err error
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

	logger.Debug("clxmax invoked", "version", cliversion.Version)

	if *showVersion {
		fmt.Fprintln(stdout, cliversion.Line("clxmax"))
		return 0
	}

	fmt.Fprintln(stdout, phaseMessage)
	return 0
}
