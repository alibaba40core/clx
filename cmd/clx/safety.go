package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/policy"
)

func runSafety(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printSafetyHelp(stderr)
		return 2
	}

	switch args[0] {
	case "show":
		return runSafetyShow(args[1:], stdout, stderr)
	case "set":
		return runSafetySet(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printSafetyHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "safety: unknown subcommand %q\n", args[0])
		printSafetyHelp(stderr)
		return 2
	}
}

func printSafetyHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx safety show [--config path]")
	fmt.Fprintln(w, "  clx safety set mode=<low|medium|high|custom> [--config path]")
	fmt.Fprintln(w, "  clx safety set require_confirmation=<true|false> [--config path]")
	fmt.Fprintln(w, "  clx safety set dry_run=<true|false> [--config path]")
	fmt.Fprintln(w, "  clx safety set explain=<true|false> [--config path]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Safety mode controls what happens after a command is classified by risk:")
	fmt.Fprintln(w, "  low     low/medium risk run; high risk confirm")
	fmt.Fprintln(w, "  medium  low risk run; medium/high explain + confirm (default)")
	fmt.Fprintln(w, "  high    low explain + confirm; medium/high preview + explain + confirm;")
	fmt.Fprintln(w, "          enforces policy allow-list (see clx policy allow/list/rm)")
	fmt.Fprintln(w, "  custom  use require_confirmation, dry_run, and explain toggles globally")
}

func runSafetyShow(args []string, stdout, stderr io.Writer) int {
	configPath, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	for _, line := range config.SafetyShowLines(cfg) {
		fmt.Fprintln(stdout, line)
	}
	return 0
}

func runSafetySet(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "safety set: key=value required")
		return 2
	}
	configPath, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	// Re-parse: parseConfigPathFlag consumes --config; find the key=value arg.
	key, value, ok := parseSafetySetArg(args)
	if !ok {
		fmt.Fprintln(stderr, "safety set: expected key=value (e.g. mode=medium)")
		return 2
	}

	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}

	key = strings.ToLower(strings.TrimSpace(key))
	switch key {
	case "mode":
		if err := config.ApplySafetyMode(&cfg, value); err != nil {
			fmt.Fprintf(stderr, "safety set: %v\n", err)
			return 1
		}
	case "require_confirmation", "dry_run", "explain":
		b, err := parseBoolValue(value)
		if err != nil {
			fmt.Fprintf(stderr, "safety set: %v\n", err)
			return 2
		}
		if err := config.ApplySafetyCustomToggle(&cfg, key, b); err != nil {
			fmt.Fprintf(stderr, "safety set: %v\n", err)
			return 1
		}
	default:
		fmt.Fprintf(stderr, "safety set: unknown key %q\n", key)
		return 2
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(stderr, "safety set: %v\n", err)
		return 1
	}
	if err := saveConfig(context.Background(), configPath, cfg, stderr); err != nil {
		fmt.Fprintf(stderr, "safety set: %v\n", err)
		return 1
	}
	ctx := context.Background()
	if strings.EqualFold(cfg.Safety.Mode, "high") {
		if err := policy.EnsureHighDefaults(ctx); err != nil {
			fmt.Fprintf(stderr, "policy defaults: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "policy: seeded allow-list defaults for high mode (git, docker, npm)")
		fmt.Fprintln(stdout, "extend with: clx policy allow <verb>")
	}
	fmt.Fprintf(stdout, "safety mode: %s\n", cfg.Safety.Mode)
	return 0
}

// parseSafetySetArg finds the first non-flag argument as key=value.
func parseSafetySetArg(args []string) (key, value string, ok bool) {
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			continue
		}
		eq := strings.IndexByte(a, '=')
		if eq <= 0 {
			continue
		}
		return a[:eq], a[eq+1:], true
	}
	return "", "", false
}

func parseBoolValue(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "1":
		return true, nil
	case "false", "no", "0":
		return false, nil
	default:
		return false, fmt.Errorf("expected true or false, got %q", s)
	}
}
