package providers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
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

	writePlatformContext(&b, req.Profile)

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

// writePlatformContext renders the machine profile lines shared by the intent
// and command-generation prompts so the model targets the correct OS/shell.
func writePlatformContext(b *strings.Builder, profile environment.SystemProfile) {
	osLine := profile.OS
	if v := strings.TrimSpace(profile.OSVersion); v != "" {
		osLine = osLine + " " + v
	}
	shellLine := profile.Shell
	if v := strings.TrimSpace(profile.ShellVersion); v != "" {
		shellLine = shellLine + " " + v
	}

	fmt.Fprintf(b, "OS: %s\nShell: %s\nTerminal: %s", osLine, shellLine, profile.Terminal)
	if profile.WSLEnabled {
		b.WriteString("\nWSL: available")
	}
	if pkgs := dedupeLower(profile.PackageManagers); len(pkgs) > 0 {
		fmt.Fprintf(b, "\nPackage managers: %s", strings.Join(pkgs, ", "))
	}
	if strings.EqualFold(profile.OS, "windows") {
		b.WriteString("\nPath hint: use \".\" for the current directory, not \"/\".")
	}
}

// dedupeLower lowercases, trims, and de-duplicates a string slice preserving order.
func dedupeLower(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		key := strings.ToLower(strings.TrimSpace(s))
		if key == "" {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
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
