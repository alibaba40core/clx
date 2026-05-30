package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/policy"
)

func runPolicy(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printPolicyHelp(stderr)
		return 2
	}
	switch args[0] {
	case "allow":
		return runPolicyAllow(args[1:], stdout, stderr)
	case "show":
		return runPolicyShow(args[1:], stdout, stderr)
	case "list":
		return runPolicyList(args[1:], stdout, stderr)
	case "rm", "remove":
		return runPolicyRm(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printPolicyHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "policy: unknown subcommand %q\n", args[0])
		printPolicyHelp(stderr)
		return 2
	}
}

func printPolicyHelp(w io.Writer) {
	fmt.Fprintln(w, "CLX policy — block list, allow list, and access level")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx policy show [--config path]   Full policy summary (recommended)")
	fmt.Fprintln(w, "  clx policy allow <verb> [--config path]")
	fmt.Fprintln(w, "  clx policy list [--config path]     Allowed verbs only (see policy show)")
	fmt.Fprintln(w, "  clx policy rm <verb> [--config path]")
	fmt.Fprintln(w, "  clx policy help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Blocked patterns in policy.yaml always apply at exec time.")
	fmt.Fprintln(w, "Allowed verbs are an opt-in argv[0] list enforced only when safety.mode=high.")
	fmt.Fprintln(w, "When the allow list is empty, all verbs are permitted except blocked patterns.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  clx policy show")
	fmt.Fprintln(w, "  clx safety set mode=high")
	fmt.Fprintln(w, "  clx policy allow pwd")
}

func runPolicyAllow(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "policy allow: verb required")
		return 2
	}
	verb := args[0]
	configPath, code := parseConfigPathFlag(args[1:], stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	if _, err := config.Bootstrap(ctx); err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}
	if err := policy.AddAllowedVerb(ctx, verb); err != nil {
		fmt.Fprintf(stderr, "policy allow: %v\n", err)
		return 1
	}
	path, _ := config.PolicyPath()
	if path != "" {
		fmt.Fprintf(stdout, "allowed verb %q (saved to %s)\n", verb, path)
	} else {
		fmt.Fprintf(stdout, "allowed verb %q\n", verb)
	}
	if cfg.Safety.Mode != "high" {
		fmt.Fprintln(stderr, "note: allowed verbs are only enforced when safety.mode=high")
	}
	return 0
}

func runPolicyShow(args []string, stdout, stderr io.Writer) int {
	configPath, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	if _, err := config.Bootstrap(ctx); err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}
	snap, err := policy.LoadSnapshot(ctx, cfg.Safety.Mode)
	if err != nil {
		fmt.Fprintf(stderr, "policy show: %v\n", err)
		return 1
	}
	printPolicySnapshot(stdout, snap)
	return 0
}

func printPolicySnapshot(w io.Writer, snap policy.Snapshot) {
	if snap.Path != "" {
		fmt.Fprintf(w, "Policy file: %s\n", snap.Path)
	} else {
		fmt.Fprintln(w, "Policy file: (not configured)")
	}
	if !snap.FileExists {
		fmt.Fprintln(w, "  (file missing — no block or allow rules loaded)")
	}
	fmt.Fprintf(w, "Safety mode: %s (from config.yaml)\n", snap.SafetyMode)
	fmt.Fprintf(w, "Access level: %s\n", snap.AccessLevel.String())
	switch snap.AccessLevel {
	case policy.AccessSafe:
		fmt.Fprintln(w, "  → all execution denied; --explain still shows translations")
	case policy.AccessModerate:
		fmt.Fprintln(w, "  → only low-risk commands may run at exec time")
	default:
		fmt.Fprintln(w, "  → commands allowed if they pass block/allow lists")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Blocked patterns (always enforced at exec):")
	if len(snap.Blocked) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, p := range snap.Blocked {
			fmt.Fprintf(w, "  - %s\n", p)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Allowed verbs (argv[0] whitelist for high safety only):")
	if len(snap.Allowed) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, v := range snap.Allowed {
			fmt.Fprintf(w, "  - %s\n", v)
		}
	}
	if policy.AllowListActive(snap.SafetyMode, snap.Allowed) {
		fmt.Fprintln(w, "  → enforced: only verbs above may run (plus block list)")
	} else if len(snap.Allowed) > 0 {
		fmt.Fprintf(w, "  → stored but not enforced (safety.mode=%s; use high to activate)\n", snap.SafetyMode)
	} else {
		fmt.Fprintln(w, "  → not enforced: all verbs OK except blocked patterns")
	}
}

func runPolicyList(args []string, stdout, stderr io.Writer) int {
	configPath, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	list, err := policy.ListAllowedVerbs(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "policy list: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Allowed verbs (not blocked — permitted argv[0] in high safety):")
	if len(list) == 0 {
		fmt.Fprintln(stdout, "  (none)")
		if cfg.Safety.Mode == "high" {
			fmt.Fprintln(stdout, "  high mode with empty list: all verbs OK except block list")
			fmt.Fprintln(stdout, "  seed defaults: clx safety set mode=high")
		}
		fmt.Fprintln(stdout, "Run `clx policy show` for blocked patterns and enforcement status.")
		return 0
	}
	for _, v := range list {
		fmt.Fprintf(stdout, "  - %s\n", v)
	}
	if policy.AllowListActive(cfg.Safety.Mode, list) {
		fmt.Fprintln(stdout, "Enforced now (safety.mode=high).")
	} else {
		fmt.Fprintf(stdout, "Not enforced (safety.mode=%s).\n", cfg.Safety.Mode)
	}
	fmt.Fprintln(stdout, "Run `clx policy show` for blocked patterns and full summary.")
	return 0
}

func runPolicyRm(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "policy rm: verb required")
		return 2
	}
	verb := args[0]
	_, code := parseConfigPathFlag(args[1:], stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	if _, err := config.Bootstrap(ctx); err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}
	if err := policy.RemoveAllowedVerb(ctx, verb); err != nil {
		if errors.Is(err, policy.ErrVerbNotFound) {
			fmt.Fprintf(stderr, "policy rm: %v\n", err)
			return 1
		}
		fmt.Fprintf(stderr, "policy rm: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "removed allowed verb %q\n", verb)
	return 0
}
