package providers

import (
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/generator"
)

const maxExplainUserBytes = 8 * 1024

// explainSystemPrompt is the global explain prompt (D1: one prompt for all domains).
const explainSystemPrompt = `Describe what the native shell command does in one or two plain sentences.
Use simple language. No markdown, no code fences, no shell metacharacters in your answer.`

// BuildExplainPrompt assembles system and user messages for Provider.Explain.
func BuildExplainPrompt(gen generator.GeneratedCommand, profile environment.SystemProfile) (system, user string, err error) {
	system = explainSystemPrompt

	var b strings.Builder
	fmt.Fprintf(&b, "OS: %s\n", profile.OS)
	shell := gen.Shell
	if shell == "" {
		shell = profile.Shell
	}
	fmt.Fprintf(&b, "Shell: %s\n", shell)
	if intent := strings.TrimSpace(gen.Intent); intent != "" {
		fmt.Fprintf(&b, "Intent: %s\n", intent)
	}
	b.WriteString("Command argv: ")
	for i, arg := range gen.Argv {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(executor.Redact(arg))
	}
	if len(gen.Argv) == 0 && gen.Command != "" {
		b.WriteString(executor.Redact(gen.Command))
	}

	user = b.String()
	if len(user) > maxExplainUserBytes {
		return "", "", fmt.Errorf("explain prompt too large: %d bytes (cap %d)", len(user), maxExplainUserBytes)
	}
	return system, user, nil
}
