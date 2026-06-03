package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/generator"
)

// maxCommandTools caps how many detected tools are listed in the command prompt
// to keep the user message under maxPromptUserBytes.
const maxCommandTools = 40

// CommandGenerator is the optional capability for providers that can synthesize
// a full command for the active platform when no rule or cached intent matches.
//
// Security contract: the returned command is UNTRUSTED. Callers MUST validate
// via executor.ValidateGeneratedArgv or executor.ValidateCommandChain before exec.
type CommandGenerator interface {
	GenerateCommand(ctx context.Context, req CommandRequest) (*CommandResponse, error)
}

// CommandRequest carries everything the command prompt/schema builders need.
type CommandRequest struct {
	RawInput string
	Profile  environment.SystemProfile
	// Feedback, when set, tells the model a prior response failed CLX validation.
	Feedback string
}

// CommandResponse is a provider-generated command before validation/gating.
// Use Chain for multi-stage commands, or Argv for a single command.
type CommandResponse struct {
	Argv        []string
	Chain       *generator.CommandChain
	Shell       string
	Explanation string
	Confidence  float64
}

// HasChain reports whether the response is a multi-stage chain.
func (r *CommandResponse) HasChain() bool {
	return r != nil && r.Chain != nil && len(r.Chain.Stages) >= 2
}

const commandSystemPrompt = `/no_think
You are CLX, a command generator. Map the user's request to shell command(s) for the platform described in the user message.

Rules:
- Prefer "chain" when the task needs pipe or sequential composition (filtering, grep after list, sorting, etc.).
- chain.stages: array of stages; each stage has "tokens": [{"value":"...","expr":false}].
- Use "expr":true only for short scriptblock/predicate tokens (e.g. Where-Object filter body). Keep expr tokens under 200 chars; split complex logic across stages.
- connectors: array of "pipe" or "and" between stages (length = stages-1). Use "pipe" for filtering; "and" for sequential success-only steps.
- Do NOT put |, &&, or ; inside token values — CLX inserts connectors.
- Never use placeholders like URL, file, or . as a stand-in path; pick sensible defaults from the request.
- For bulk rename, pipe Get-ChildItem to Rename-Item; never use semicolon between stages.
- Never pipe to bare $_ or { scriptblock } alone — always use Where-Object or ForEach-Object as the stage command, with the predicate in an expr:true token.
- For a single simple command, use flat "argv" and set "chain" to null.
- Use programs on the platform; set "shell" (cmd, powershell, bash, sh).
- Set "explanation" and "confidence" (0-1).
Respond with JSON only.

Example (largest files, PowerShell chain):
{"argv":[],"chain":{"stages":[{"tokens":[{"value":"Get-ChildItem","expr":false},{"value":".","expr":false},{"value":"-File","expr":false},{"value":"-Recurse","expr":false}]},{"tokens":[{"value":"Sort-Object","expr":false},{"value":"Length","expr":false},{"value":"-Descending","expr":false}]},{"tokens":[{"value":"Select-Object","expr":false},{"value":"-First","expr":false},{"value":"10","expr":false}]}],"connectors":["pipe","pipe"]},"shell":"powershell","explanation":"List 10 largest files","confidence":0.9}

Example (CPU and memory, PowerShell chain):
{"argv":[],"chain":{"stages":[{"tokens":[{"value":"Get-CimInstance","expr":false},{"value":"Win32_OperatingSystem","expr":false}]},{"tokens":[{"value":"Select-Object","expr":false},{"value":"TotalVisibleMemorySize","expr":false},{"value":"FreePhysicalMemory","expr":false}]}],"connectors":["pipe"]},"shell":"powershell","explanation":"Show memory stats","confidence":0.9}

Example (filter with Where-Object, PowerShell chain):
{"argv":[],"chain":{"stages":[{"tokens":[{"value":"Get-NetTCPConnection","expr":false},{"value":"-LocalPort","expr":false},{"value":"443","expr":false}]},{"tokens":[{"value":"Where-Object","expr":false},{"value":"{$_.State -eq 'Listen'}","expr":true}]}],"connectors":["pipe"]},"shell":"powershell","explanation":"Show listeners on port 443","confidence":0.9}

Example (simple single command):
{"argv":["git","status"],"chain":null,"shell":"powershell","explanation":"Show git status","confidence":0.95}`

// BuildCommandPrompt assembles the system and user messages for AI command generation.
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
	if fb := strings.TrimSpace(req.Feedback); fb != "" {
		b.WriteString("\n\nPrevious attempt was rejected: ")
		b.WriteString(executor.Redact(fb))
		b.WriteString(". Return a valid chain or argv without shell metacharacters in tokens.")
	}
	return b.String()
}

var chainTokenSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []any{"value", "expr"},
	"properties": map[string]any{
		"value": map[string]any{"type": "string"},
		"expr":  map[string]any{"type": "boolean"},
	},
}

var chainStageSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []any{"tokens"},
	"properties": map[string]any{
		"tokens": map[string]any{"type": "array", "items": chainTokenSchema},
	},
}

// BuildCommandSchema returns JSON Schema for command generation (always allows chains).
func BuildCommandSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"argv", "shell", "explanation", "confidence"},
		"properties": map[string]any{
			"argv": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"chain": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"stages":     map[string]any{"type": "array", "items": chainStageSchema},
					"connectors": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []any{"stages", "connectors"},
			},
			"shell":       map[string]any{"type": "string"},
			"explanation": map[string]any{"type": "string"},
			"confidence":  map[string]any{"type": "number"},
		},
	}
}

type parsedChainToken struct {
	Value string `json:"value"`
	Expr  bool   `json:"expr"`
}

type parsedChainStage struct {
	Tokens []parsedChainToken `json:"tokens"`
}

type parsedChain struct {
	Stages     []parsedChainStage `json:"stages"`
	Connectors []string           `json:"connectors"`
}

type parsedCommand struct {
	Argv        []string     `json:"argv"`
	Chain       *parsedChain `json:"chain"`
	Shell       string       `json:"shell"`
	Explanation string       `json:"explanation"`
	Confidence  float64      `json:"confidence"`
}

// ExtractJSONObject returns the first {...} span in model text (handles markdown fences).
func ExtractJSONObject(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "{") {
		return content
	}
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	return content
}

// ParseCommandContent decodes a model message into a CommandResponse.
func ParseCommandContent(content string) (*CommandResponse, error) {
	raw := strings.TrimSpace(content)
	candidates := []string{raw, ExtractJSONObject(raw)}
	var parsed parsedCommand
	var lastErr error
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		lastErr = json.Unmarshal([]byte(candidate), &parsed)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return nil, ErrInvalidResp
	}
	if chain := parseChain(parsed.Chain); chain != nil && len(chain.Stages) >= 2 {
		return &CommandResponse{
			Chain:       chain,
			Shell:       strings.TrimSpace(parsed.Shell),
			Explanation: strings.TrimSpace(parsed.Explanation),
			Confidence:  parsed.Confidence,
		}, nil
	}
	cleaned := make([]string, 0, len(parsed.Argv))
	for _, tok := range parsed.Argv {
		if tok != "" {
			cleaned = append(cleaned, tok)
		}
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

func parseChain(pc *parsedChain) *generator.CommandChain {
	if pc == nil || len(pc.Stages) < 2 {
		return nil
	}
	stages := make([]generator.ChainStage, 0, len(pc.Stages))
	for _, st := range pc.Stages {
		toks := make([]generator.ChainToken, 0, len(st.Tokens))
		for _, t := range st.Tokens {
			v := strings.TrimSpace(t.Value)
			if v == "" {
				continue
			}
			toks = append(toks, generator.ChainToken{Value: v, Expr: t.Expr})
		}
		if len(toks) == 0 {
			return nil
		}
		stages = append(stages, generator.ChainStage{Tokens: toks})
	}
	if len(stages) < 2 {
		return nil
	}
	conns := make([]generator.ChainConnector, len(stages)-1)
	for i := 0; i < len(stages)-1; i++ {
		conns[i] = generator.ChainPipe
		if i < len(pc.Connectors) {
			switch strings.ToLower(strings.TrimSpace(pc.Connectors[i])) {
			case "and", "&&":
				conns[i] = generator.ChainAnd
			default:
				conns[i] = generator.ChainPipe
			}
		}
	}
	return &generator.CommandChain{Stages: stages, Connectors: conns}
}

// BuildOpenAICommandSchema returns a strict-mode-compatible schema (optional chain as null).
func BuildOpenAICommandSchema() map[string]any {
	chainObject := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"stages":     map[string]any{"type": "array", "items": chainStageSchema},
			"connectors": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required": []any{"stages", "connectors"},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"argv", "chain", "shell", "explanation", "confidence"},
		"properties": map[string]any{
			"argv": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"chain": map[string]any{
				"anyOf": []any{
					chainObject,
					map[string]any{"type": "null"},
				},
			},
			"shell":       map[string]any{"type": "string"},
			"explanation": map[string]any{"type": "string"},
			"confidence":  map[string]any{"type": "number"},
		},
	}
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

// ChainFromArgv splits argv on connector tokens into a CommandChain.
func ChainFromArgv(argv []string) *generator.CommandChain {
	return generator.ChainFromArgv(argv)
}

// ArgvHasChainConnector reports whether argv contains a chain connector token.
func ArgvHasChainConnector(argv []string) bool {
	return generator.ArgvHasChainConnector(argv)
}
