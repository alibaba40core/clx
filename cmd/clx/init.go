package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alibaba40core/clx/internal/config"
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

	provider := promptChoice(stdout, in, "AI provider", []string{
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
		fmt.Fprintf(stdout, "Set %s now? [y/N]: ", secretPath)
		if yes, _ := readYesNo(in); yes {
			val, rerr := readSecretValue(os.Stdin, stderr, secretPath)
			if rerr != nil {
				fmt.Fprintf(stderr, "secret: %v\n", rerr)
			} else if val != "" {
				_ = config.SetByPath(&cfg, secretPath, val)
			}
		}
	}

	safety := promptChoice(stdout, in, "Safety mode", []string{
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

	fmt.Fprint(stdout, "Install shell hook snippet (explain-only)? [y/N]: ")
	if yes, _ := readYesNo(in); yes {
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
}

func promptChoice(stdout io.Writer, in *bufio.Reader, title string, options []string, defaultIdx int) int {
	fmt.Fprintf(stdout, "\n%s:\n", title)
	for i, opt := range options {
		mark := " "
		if i == defaultIdx {
			mark = "*"
		}
		fmt.Fprintf(stdout, "  %s %d) %s\n", mark, i+1, opt)
	}
	fmt.Fprintf(stdout, "Choice [%d]: ", defaultIdx+1)
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultIdx
	}
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(options) {
		return defaultIdx
	}
	return n - 1
}

func readYesNo(in *bufio.Reader) (bool, error) {
	line, err := in.ReadString('\n')
	if err != nil && line == "" {
		return false, err
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes", nil
}
