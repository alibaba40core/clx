package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/policy"
)

func runInit(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		printInitHelp(stdout)
		return 0
	}

	ctx := context.Background()
	in := bufio.NewReader(os.Stdin)

	fmt.Fprintln(stdout, "CLX setup wizard")
	fmt.Fprintln(stdout)

	result, err := config.Bootstrap(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}
	if result.WroteConfig {
		fmt.Fprintln(stdout, "Created ~/.clx/ with default config and policy.")
	} else {
		fmt.Fprintln(stdout, "Using existing ~/.clx/ configuration.")
	}

	cfgPath, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(stderr, "config path: %v\n", err)
		return 1
	}
	cfg, err := config.Load(ctx, cfgPath)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}

	provider := promptChoice(stdout, stderr, in, "AI provider", []string{
		"ollama (local)",
		"openai",
		"gemini",
		"skip (rules only)",
	}, 0)
	switch provider {
	case 1:
		cfg.Provider = "openai"
		cfg.Providers.Primary = "openai"
	case 2:
		cfg.Provider = "gemini"
		cfg.Providers.Primary = "gemini"
	case 3:
		// Rules only: disable the AI provider entirely so CLX never depends on
		// Ollama or any LLM. Rule-backed commands still work.
		cfg.Provider = "none"
		cfg.Providers.Primary = "none"
		cfg.Providers.Fallback = ""
		cfg.Features.AICommandGeneration = false
	default:
		cfg.Provider = "ollama"
		cfg.Providers.Primary = "ollama"
	}

	if provider == 1 || provider == 2 {
		secretPath := "providers.openai.api_key"
		if provider == 2 {
			secretPath = "providers.gemini.api_key"
		}
		// Never use a visible [y/N] line for secrets — pasted keys would echo to the terminal.
		fmt.Fprintf(stdout, "\n%s (hidden prompt on stderr; press Enter with no input to skip)\n", secretPath)
		val, rerr := readSecretValue(os.Stdin, stderr, secretPath)
		if rerr != nil {
			fmt.Fprintf(stderr, "secret: %v\n", rerr)
		} else if val != "" {
			if err := config.SetByPath(&cfg, secretPath, val); err != nil {
				fmt.Fprintf(stderr, "secret: %v\n", err)
			}
		}
	}

	safety := promptChoice(stdout, stderr, in, "Safety mode", []string{
		"medium (default)",
		"low",
		"high",
	}, 0)
	switch safety {
	case 1:
		_ = config.ApplySafetyMode(&cfg, "low")
	case 2:
		_ = config.ApplySafetyMode(&cfg, "high")
	default:
		_ = config.ApplySafetyMode(&cfg, "medium")
	}

	if promptYesNo(stdout, stderr, in, "Install shell hook snippet (explain-only)? [y/N]: ", true) {
		cfg.Execution.ShellIntegration = true
		fmt.Fprintln(stdout, string(config.EmbeddedShellIntegration()))
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(stderr, "config: %v\n", err)
		return 1
	}
	if err := config.Save(ctx, cfgPath, cfg); err != nil {
		fmt.Fprintf(stderr, "save config: %v\n", err)
		return 1
	}
	if safety == 2 {
		if err := policy.EnsureHighDefaults(ctx); err != nil {
			fmt.Fprintf(stderr, "policy defaults: %v\n", err)
			return 1
		}
	}
	fmt.Fprintln(stdout, "Configuration saved.")

	fmt.Fprintln(stdout, "\nNext: run `clx doctor` to detect tools and write your system profile.")
	return 0
}

func printInitHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx init")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Interactive first-run wizard: bootstrap ~/.clx, provider, safety mode,")
	fmt.Fprintln(w, "optional shell hook instructions, then suggests clx doctor.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "API keys are read with hidden terminal input (same as clx config set).")
	fmt.Fprintln(w, "Do not paste secrets on yes/no prompts.")
}

