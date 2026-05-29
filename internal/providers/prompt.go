package providers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alibaba40core/clx/internal/executor"
)

const (
	maxKnownIntents    = 256
	maxPromptUserBytes = 8 * 1024
)

// systemPrompt is the global resolution prompt (D1: one prompt for all domains).
// /no_think disables Qwen3 chain-of-thought for faster JSON-only extraction.
const systemPrompt = `/no_think
You map the user's shell or natural-language request to exactly one intent from the allowed list in the user message. Extract only parameters declared for that intent. Respond with JSON only: {"intent":"<name>","params":{...},"confidence":<0-1>}.`

// toolAllowlist limits which detected tools are included in grounding context.
var toolAllowlist = []string{
	"curl", "docker", "fd", "find", "git", "grep", "kubectl", "ls", "rg", "ssh",
}

// BuildPrompt assembles the system and user messages for intent resolution.
// Output is deterministic for identical IntentRequest values.
func BuildPrompt(req IntentRequest) (system, user string, err error) {
	if len(req.KnownIntents) > maxKnownIntents {
		return "", "", fmt.Errorf("too many known intents: %d (cap %d)", len(req.KnownIntents), maxKnownIntents)
	}

	intents := append([]string(nil), req.KnownIntents...)
	sort.Strings(intents)

	tools := filterTools(req.Profile.AvailableTools)
	user = buildUserMessage(req, intents, tools)

	if len(user) > maxPromptUserBytes {
		return "", "", fmt.Errorf("prompt too large: %d bytes (cap %d)", len(user), maxPromptUserBytes)
	}

	return systemPrompt, user, nil
}

func buildUserMessage(req IntentRequest, intents, tools []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "OS: %s\nShell: %s\nTerminal: %s",
		req.Profile.OS, req.Profile.Shell, req.Profile.Terminal)
	if strings.EqualFold(req.Profile.OS, "windows") {
		b.WriteString("\nPath hint: use \".\" for the current directory, not \"/\".")
	}

	b.WriteString("\n\nAvailable tools: ")
	if len(tools) == 0 {
		b.WriteString("(none detected from allowlist)")
	} else {
		b.WriteString(strings.Join(tools, ", "))
	}

	b.WriteString("\n\nAllowed intents: ")
	if len(intents) == 0 {
		b.WriteString("(none)")
	} else {
		b.WriteString(strings.Join(intents, ", "))
	}

	if len(req.RuleParams) > 0 {
		b.WriteString("\n\nIntent parameters (use only keys for the chosen intent):")
		for _, name := range intents {
			ps := req.RuleParams[name]
			if len(ps) == 0 {
				fmt.Fprintf(&b, "\n- %s: none", name)
			} else {
				fmt.Fprintf(&b, "\n- %s: %s", name, strings.Join(ps, ", "))
			}
		}
	}

	b.WriteString("\n\nInput: ")
	b.WriteString(executor.Redact(req.RawInput))

	return b.String()
}

func filterTools(available []string) []string {
	if len(available) == 0 {
		return nil
	}
	allow := make(map[string]struct{}, len(toolAllowlist))
	for _, t := range toolAllowlist {
		allow[t] = struct{}{}
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(available))
	for _, t := range available {
		key := strings.ToLower(strings.TrimSpace(t))
		if key == "" {
			continue
		}
		if _, ok := allow[key]; !ok {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
