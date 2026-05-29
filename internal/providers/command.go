package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
)

// maxCommandTools caps how many detected tools are listed in the command prompt
// to keep the user message under maxPromptUserBytes.
const maxCommandTools = 40

// CommandGenerator is the optional capability for providers that can synthesize
// a full command (as argv tokens) for the active platform when no rule or
// cached intent matches. It is intentionally separate from Provider so existing
// implementations and test fakes are not forced to implement it.
//
// Security contract: the returned argv is UNTRUSTED. Callers MUST run it through
// executor.ValidateGeneratedArgv, internal/risk, and internal/policy before exec,
// and MUST execute it argv-only (never as an interpolated shell string).
type CommandGenerator interface {
	GenerateCommand(ctx context.Context, req CommandRequest) (*CommandResponse, error)
}

// CommandRequest carries everything the command prompt/schema builders need.
type CommandRequest struct {
	RawInput string
	Profile  environment.SystemProfile
}

// CommandResponse is a provider-generated command before validation/gating.
// Argv is the tokenized command (program + arguments), already split — it must
// not contain shell operators; those are rejected downstream.
type CommandResponse struct {
	Argv        []string
	Shell       string // target shell hint: cmd|powershell|pwsh|bash|sh|zsh
	Explanation string
	Confidence  float64
}

// commandSystemPrompt instructs the model to emit a single, safe, platform-correct
// command as discrete argv tokens. The hard safety guarantees (no shell operators,
// argv-only exec, risk/policy gating) are enforced in Go regardless of this text.
const commandSystemPrompt = `/no_think
You are CLX, a command generator. Map the user's request to exactly ONE shell command for the platform described in the user message.

Rules:
- Output the command as "argv": an array of tokens already split into the program and each argument. Example: ["git","status"] not ["git status"].
- Generate exactly ONE command. Do NOT chain commands or use shell operators: no pipes (|), redirects (> <), &&, ||, ;, backticks, $(...), or %VAR% expansion. If the task needs multiple steps, pick the single most useful step.
- Use a program that exists on the platform (prefer the listed available tools and the native shell).
- Use the correct syntax for the given OS and shell (e.g. "dir" on Windows cmd, "ls" on POSIX).
- Set "shell" to the shell the command targets (cmd, powershell, bash, or sh).
- Set "explanation" to a short one-line description.
- Set "confidence" between 0 and 1.
Respond with JSON only: {"argv":["..."],"shell":"<shell>","explanation":"<text>","confidence":<0-1>}.`

// BuildCommandPrompt assembles the system and user messages for AI command
// generation. Output is deterministic for identical CommandRequest values.
func BuildCommandPrompt(req CommandRequest) (system, user string, err error) {
	user = buildCommandUserMessage(req)
	if len(user) > maxPromptUserBytes {
		return "", "", fmt.Errorf("command prompt too large: %d bytes (cap %d)", len(user), maxPromptUserBytes)
	}
	return commandSystemPrompt, user, nil
}

func buildCommandUserMessage(req CommandRequest) string {
	var b strings.Builder

	writePlatformContext(&b, req.Profile)

	tools := dedupeLower(req.Profile.AvailableTools)
	if len(tools) > maxCommandTools {
		tools = tools[:maxCommandTools]
	}
	b.WriteString("\n\nAvailable tools: ")
	if len(tools) == 0 {
		b.WriteString("(none detected)")
	} else {
		b.WriteString(strings.Join(tools, ", "))
	}

	b.WriteString("\n\nRequest: ")
	b.WriteString(executor.Redact(req.RawInput))

	return b.String()
}

// BuildCommandSchema returns a JSON Schema constraining the command response.
// It is strict-mode compatible (additionalProperties:false, all keys required)
// so it works for both Ollama format and OpenAI structured outputs. Numeric and
// array bounds are validated in Go (executor.ValidateGeneratedArgv), not here,
// because the strict subset does not support them on all providers.
func BuildCommandSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"argv", "shell", "explanation", "confidence"},
		"properties": map[string]any{
			"argv": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"shell":       map[string]any{"type": "string"},
			"explanation": map[string]any{"type": "string"},
			"confidence":  map[string]any{"type": "number"},
		},
	}
}

// parsedCommand is the JSON shape expected inside the model's message content.
type parsedCommand struct {
	Argv        []string `json:"argv"`
	Shell       string   `json:"shell"`
	Explanation string   `json:"explanation"`
	Confidence  float64  `json:"confidence"`
}

// ParseCommandContent decodes a model message into a CommandResponse. It returns
// ErrInvalidResp for malformed JSON and ErrNoMatch when no argv is produced.
// It does NOT validate command safety — that is the caller's responsibility.
func ParseCommandContent(content string) (*CommandResponse, error) {
	var parsed parsedCommand
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, ErrInvalidResp
	}
	cleaned := make([]string, 0, len(parsed.Argv))
	for _, tok := range parsed.Argv {
		if tok == "" {
			continue
		}
		cleaned = append(cleaned, tok)
	}
	if len(cleaned) == 0 {
		return nil, ErrNoMatch
	}
	return &CommandResponse{
		Argv:        cleaned,
		Shell:       strings.TrimSpace(parsed.Shell),
		Explanation: strings.TrimSpace(parsed.Explanation),
		Confidence:  parsed.Confidence,
	}, nil
}

// BuildOpenAICommandResponseFormat wraps the command schema for OpenAI chat.
func BuildOpenAICommandResponseFormat(schema map[string]any) map[string]any {
	return map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   "clx_command",
			"strict": true,
			"schema": schema,
		},
	}
}
