package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/executor"
)

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printConfigHelp(stderr)
		return 2
	}

	switch args[0] {
	case "show":
		return runConfigShow(args[1:], stdout, stderr)
	case "get":
		return runConfigGet(args[1:], stdout, stderr)
	case "set":
		return runConfigSet(args[1:], stdout, stderr)
	case "provider":
		return runConfigProvider(args[1:], stdout, stderr)
	case "encrypt-secrets":
		return runConfigEncryptSecrets(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printConfigHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "config: unknown subcommand %q\n", args[0])
		printConfigHelp(stderr)
		return 2
	}
}

func printConfigHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx config show [--config path]")
	fmt.Fprintln(w, "  clx config get <path> [--config path]")
	fmt.Fprintln(w, "  clx config set <path> <value> [--config path]")
	fmt.Fprintln(w, "  clx config set <secret-path> [--stdin] [--config path]")
	fmt.Fprintln(w, "  clx config set <path> --stdin [--config path]")
	fmt.Fprintln(w, "  clx config provider list")
	fmt.Fprintln(w, "  clx config provider use <ollama|openai|azure|gemini> [--config path]")
	fmt.Fprintln(w, "  clx config encrypt-secrets [--config path]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Provider paths: provider, model, providers.primary, providers.fallback,")
	fmt.Fprintln(w, "  providers.timeout, providers.ollama.host, providers.ollama.model,")
	fmt.Fprintln(w, "  providers.openai.api_key, providers.openai.model,")
	fmt.Fprintln(w, "  providers.azure.endpoint, providers.azure.api_key, providers.azure.deployment,")
	fmt.Fprintln(w, "  providers.gemini.api_key, providers.gemini.model")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Secret paths (api_key): omit the value or use --stdin for a hidden prompt.")
	fmt.Fprintln(w, "Do not pass API keys as command-line arguments.")
}

func runConfigShow(args []string, stdout, stderr io.Writer) int {
	path, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(path, stderr)
	if code != 0 {
		return code
	}
	for _, line := range config.ProviderShowLines(cfg) {
		fmt.Fprintln(stdout, line)
	}
	return 0
}

func runConfigGet(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "config get: path required")
		return 2
	}
	pathKey := args[0]
	rest := args[1:]
	configPath, code := parseConfigPathFlag(rest, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	val, err := config.GetByPath(cfg, pathKey)
	if err != nil {
		fmt.Fprintf(stderr, "config get: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, val)
	return 0
}

func runConfigSet(args []string, stdout, stderr io.Writer) int {
	configPath, useStdin, remaining, errMsg := parseConfigSetArgs(args)
	if errMsg != "" {
		fmt.Fprintln(stderr, errMsg)
		return 2
	}
	if len(remaining) < 1 {
		fmt.Fprintln(stderr, "config set: path required")
		return 2
	}
	pathKey := remaining[0]
	isSecret := config.IsSecretPath(pathKey)

	if len(remaining) >= 2 && isSecret {
		fmt.Fprintln(stderr, "config set: secret paths cannot receive a value on the command line; omit the value or use --stdin for a hidden prompt")
		return 2
	}

	var value string
	var err error
	switch {
	case isSecret && useStdin:
		value, err = readPlainStdin(os.Stdin)
		if err != nil {
			fmt.Fprintf(stderr, "config set: read stdin: %v\n", err)
			return 1
		}
	case isSecret:
		value, err = readSecretValue(os.Stdin, stderr, pathKey)
		if err != nil {
			fmt.Fprintf(stderr, "config set: %v\n", err)
			return 1
		}
	case useStdin:
		value, err = readPlainStdin(os.Stdin)
		if err != nil {
			fmt.Fprintf(stderr, "config set: read stdin: %v\n", err)
			return 1
		}
	case len(remaining) >= 2:
		value = remaining[1]
	default:
		fmt.Fprintln(stderr, "config set: value or --stdin required")
		return 2
	}

	if isSecret && value == "" {
		fmt.Fprintln(stderr, "config set: secret value cannot be empty")
		return 1
	}

	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	if err := config.SetByPath(&cfg, pathKey, value); err != nil {
		fmt.Fprintf(stderr, "config set: %v\n", err)
		return 1
	}
	if err := saveConfig(context.Background(), configPath, cfg, stderr); err != nil {
		fmt.Fprintf(stderr, "config set: %v\n", executor.Redact(err.Error()))
		return 1
	}
	return 0
}

func runConfigProvider(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "config provider: subcommand required (list, use)")
		return 2
	}
	switch args[0] {
	case "list":
		fmt.Fprintln(stdout, "ollama")
		fmt.Fprintln(stdout, "openai")
		fmt.Fprintln(stdout, "azure")
		fmt.Fprintln(stdout, "gemini")
		return 0
	case "use":
		return runConfigProviderUse(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "config provider: unknown subcommand %q\n", args[0])
		return 2
	}
}

func runConfigProviderUse(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("config provider use", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "path to config.yaml")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "config provider use: provider name required")
		return 2
	}
	name := strings.ToLower(strings.TrimSpace(fs.Arg(0)))
	switch name {
	case "ollama", "openai", "azure", "gemini":
	default:
		fmt.Fprintf(stderr, "config provider use: invalid provider %q\n", name)
		return 2
	}

	cfg, code := loadConfigForEdit(*configPath, stderr)
	if code != 0 {
		return code
	}
	config.SetProviderActive(&cfg, name)
	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(stderr, "config provider use: %v\n", err)
		return 1
	}
	if err := saveConfig(context.Background(), *configPath, cfg, stderr); err != nil {
		fmt.Fprintf(stderr, "config provider use: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "active provider: %s\n", name)
	return 0
}

func runConfigEncryptSecrets(args []string, stdout, stderr io.Writer) int {
	path, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(path, stderr)
	if code != 0 {
		return code
	}
	if err := saveConfig(context.Background(), path, cfg, stderr); err != nil {
		fmt.Fprintf(stderr, "config encrypt-secrets: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "secrets encrypted in config")
	return 0
}

func parseConfigPathFlag(args []string, stderr io.Writer) (string, int) {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "path to config.yaml")
	if err := fs.Parse(args); err != nil {
		return "", 2
	}
	if *configPath != "" {
		return *configPath, 0
	}
	path, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(stderr, "config path: %v\n", err)
		return "", 1
	}
	return path, 0
}

func loadConfigForEdit(path string, stderr io.Writer) (config.Config, int) {
	if path == "" {
		var err error
		path, err = config.ConfigPath()
		if err != nil {
			fmt.Fprintf(stderr, "config path: %v\n", err)
			return config.Config{}, 1
		}
	}
	ctx := context.Background()
	if _, err := config.Bootstrap(ctx); err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return config.Config{}, 1
	}
	cfg, err := config.Load(ctx, path)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", executor.Redact(err.Error()))
		return config.Config{}, 1
	}
	return cfg, 0
}

func saveConfig(ctx context.Context, path string, cfg config.Config, stderr io.Writer) error {
	if path == "" {
		var err error
		path, err = config.ConfigPath()
		if err != nil {
			return err
		}
	}
	if err := config.Save(ctx, path, cfg); err != nil {
		return err
	}
	slog.Default().Debug("config saved", "path", path)
	return nil
}
