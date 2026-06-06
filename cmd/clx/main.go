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
	"github.com/alibaba40core/clx/internal/pipeline"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "doctor" {
		return runDoctor(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "config" {
		return runConfig(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "safety" {
		return runSafety(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "alias" {
		return runAlias(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "policy" {
		return runPolicy(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "init" {
		return runInit(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "cache" {
		return runCache(args[1:], stdout, stderr)
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
	providerFlag := fs.String("provider", "", "override AI provider (none, ollama, openai, gemini)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showHelp {
		printHelp(stdout)
		return 0
	}

	if *showVersion {
		fmt.Fprintln(stdout, cliversion.Line("clx"))
		return 0
	}

	if fs.NArg() == 0 {
		printHelp(stdout)
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, cfg, closer, err := initMinimal(ctx, *configPath, stderr)
	if err != nil {
		return 1
	}
	defer closer.Close()

	logger.Debug("clx invoked", "version", cliversion.Version)

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
		Explain:       *explain,
		DryRun:        *dryRun,
		Yes:           skipConfirm,
		Logger:        logger,
		Stdout:        stdout,
		Stderr:        stderr,
		ForwardedArgv: args,
	})
	if err != nil && code == 0 {
		return 1
	}
	return code
}

func runDoctor(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	refresh := fs.Bool("refresh", false, "force full re-detect for current shell (after installing tools)")
	refreshShort := fs.Bool("r", false, "alias for --refresh")
	configPath := fs.String("config", "", "path to config.yaml (default: ~/.clx/config.yaml)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, _, closer, err := initFull(ctx, *configPath, stderr)
	if err != nil {
		return 1
	}
	defer closer.Close()

	logger.Info("clx doctor started")
	if err := environment.RunDoctor(ctx, stdout, environment.DoctorOptions{
		Refresh: *refresh || *refreshShort,
	}); err != nil {
		fmt.Fprintf(stderr, "doctor: %v\n", err)
		return 1
	}
	return 0
}

// initCLX is kept for subcommands that still call it during migration.
func initCLX(ctx context.Context, configPath string, stderr io.Writer) (*slog.Logger, config.Config, io.Closer, error) {
	return initFull(ctx, configPath, stderr)
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "CLX — cross-platform command intelligence")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx [flags] <command...>")
	fmt.Fprintln(w, "  clx doctor [flags]")
	fmt.Fprintln(w, "  clx config <subcommand>")
	fmt.Fprintln(w, "  clx safety <subcommand>")
	fmt.Fprintln(w, "  clx policy <subcommand>")
	fmt.Fprintln(w, "  clx alias <subcommand>")
	fmt.Fprintln(w, "  clx cache <subcommand>")
	fmt.Fprintln(w, "  clx init")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  clx grep errors logs.txt")
	fmt.Fprintln(w, "  clx --explain pwd")
	fmt.Fprintln(w, "  clx --dry-run ls .")
	fmt.Fprintln(w, "  clx -y pwd")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  doctor [--refresh]  Detect environment and write ~/.clx/system_profile.json")
	fmt.Fprintln(w, "                      (run first-time or after installing tools / switching shells)")
	fmt.Fprintln(w, "  config              View or update config (providers, features, cache; see clx config help)")
	fmt.Fprintln(w, "  safety              Set safety mode and custom toggles (see clx safety help)")
	fmt.Fprintln(w, "  policy              Block list, access level, allow list (see clx policy help)")
	fmt.Fprintln(w, "  alias               Manage user-global command aliases (see clx alias help)")
	fmt.Fprintln(w, "  cache               Inspect or clear intent/explanation caches (see clx cache help)")
	fmt.Fprintln(w, "  init                Interactive first-run setup wizard")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --explain       Show intent and translation without executing")
	fmt.Fprintln(w, "  --dry-run       Preview command without executing (default; flag forces preview even if config disables it)")
	fmt.Fprintln(w, "  -y, --yes       Skip confirmation and execute")
	fmt.Fprintln(w, "  --version       Print version and exit")
	fmt.Fprintln(w, "  --help          Show this help")
	fmt.Fprintln(w, "  --config path   Path to config.yaml")
	fmt.Fprintln(w, "  --provider      Override AI provider (none, ollama, openai, gemini; none = rules only)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Exit codes: 0 success, 1 error, 2 flag error")
}
