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
	fmt.Fprintln(w, "CLX policy — allow-list verbs for high safety mode")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx policy allow <verb> [--config path]")
	fmt.Fprintln(w, "  clx policy list [--config path]")
	fmt.Fprintln(w, "  clx policy rm <verb> [--config path]")
	fmt.Fprintln(w, "  clx policy help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "The block list in ~/.clx/policies/policy.yaml always applies.")
	fmt.Fprintln(w, "The allowed list is enforced only when safety.mode=high; low, medium,")
	fmt.Fprintln(w, "and custom modes ignore allowed (block list still applies).")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  clx safety set mode=high")
	fmt.Fprintln(w, "  clx policy allow pwd")
	fmt.Fprintln(w, "  clx policy list")
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

func runPolicyList(args []string, stdout, stderr io.Writer) int {
	configPath, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	_ = cfg
	ctx := context.Background()
	list, err := policy.ListAllowedVerbs(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "policy list: %v\n", err)
		return 1
	}
	if len(list) == 0 {
		fmt.Fprintln(stdout, "no allowed verbs defined")
		if cfg.Safety.Mode == "high" {
			fmt.Fprintln(stdout, "high mode: run `clx safety set mode=high` to seed defaults, or `clx policy allow <verb>`")
		}
		return 0
	}
	for _, v := range list {
		fmt.Fprintln(stdout, v)
	}
	if cfg.Safety.Mode != "high" {
		fmt.Fprintln(stderr, "note: allowed verbs are only enforced when safety.mode=high")
	}
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
