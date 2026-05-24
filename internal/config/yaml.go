package config

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const maxYAMLBytes = 64 * 1024

// yamlNode is an internal tree node for the strict YAML subset.
type yamlNode struct {
	scalar   string
	children map[string]*yamlNode
}

func newYAMLNode() *yamlNode {
	return &yamlNode{children: make(map[string]*yamlNode)}
}

// Decode reads a strict-subset YAML document (scalar key:value only, 2-space indent).
func Decode(r io.Reader) (*yamlNode, error) {
	limited := io.LimitReader(r, maxYAMLBytes)
	sc := bufio.NewScanner(limited)
	sc.Buffer(make([]byte, 0, 4096), maxYAMLBytes)

	root := newYAMLNode()
	stack := []struct {
		indent int
		node   *yamlNode
	}{{indent: -1, node: root}}

	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.Contains(line, "\t") {
			return nil, fmt.Errorf("yaml line %d: tabs are not allowed", lineNum)
		}
		indent := leadingSpaces(line)
		if indent%2 != 0 {
			return nil, fmt.Errorf("yaml line %d: indent must be a multiple of 2", lineNum)
		}

		// Reject list items and flow syntax.
		if strings.HasPrefix(trimmed, "-") {
			return nil, fmt.Errorf("yaml line %d: lists are not supported", lineNum)
		}
		if strings.ContainsAny(trimmed, "[]{}&*!|>") {
			return nil, fmt.Errorf("yaml line %d: unsupported syntax", lineNum)
		}

		key, value, hasValue, err := parseKeyValue(trimmed, lineNum)
		if err != nil {
			return nil, err
		}

		for len(stack) > 1 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1].node

		if !hasValue {
			child := newYAMLNode()
			parent.children[key] = child
			stack = append(stack, struct {
				indent int
				node   *yamlNode
			}{indent: indent, node: child})
			continue
		}

		if len(stack) > 1 && indent <= stack[len(stack)-1].indent {
			return nil, fmt.Errorf("yaml line %d: unexpected scalar at indent %d", lineNum, indent)
		}
		parent.children[key] = &yamlNode{scalar: value}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	return root, nil
}

func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	if n < len(s) && s[n] == '\t' {
		return -1 // tabs not allowed; caller checks via indent%2
	}
	return n
}

func parseKeyValue(line string, lineNum int) (key, value string, hasValue bool, err error) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false, fmt.Errorf("yaml line %d: expected key: value", lineNum)
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false, fmt.Errorf("yaml line %d: empty key", lineNum)
	}
	rest := strings.TrimSpace(line[idx+1:])
	if rest == "" {
		return key, "", false, nil
	}
	return key, unquoteScalar(rest), true, nil
}

func unquoteScalar(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func (n *yamlNode) get(path ...string) (string, bool) {
	cur := n
	for i, p := range path {
		child, ok := cur.children[p]
		if !ok {
			return "", false
		}
		if i == len(path)-1 {
			return child.scalar, true
		}
		cur = child
	}
	return "", false
}

func (n *yamlNode) has(path ...string) bool {
	cur := n
	for _, p := range path {
		child, ok := cur.children[p]
		if !ok {
			return false
		}
		cur = child
	}
	return true
}

// applyNode merges a decoded YAML tree onto cfg (non-zero / present keys win).
func applyNode(cfg *Config, root *yamlNode) {
	if v, ok := root.get("provider"); ok {
		cfg.Provider = v
	}
	if v, ok := root.get("model"); ok {
		cfg.Model = v
	}
	if root.has("providers", "ollama", "host") {
		if v, ok := root.get("providers", "ollama", "host"); ok {
			cfg.Providers.Ollama.Host = v
		}
	}
	if root.has("providers", "ollama", "model") {
		if v, ok := root.get("providers", "ollama", "model"); ok {
			cfg.Providers.Ollama.Model = v
		}
	}
	if v, ok := root.get("providers", "openai", "api_key"); ok {
		cfg.Providers.OpenAI.APIKey = v
	}
	if v, ok := root.get("providers", "openai", "model"); ok {
		cfg.Providers.OpenAI.Model = v
	}
	if v, ok := root.get("providers", "azure", "endpoint"); ok {
		cfg.Providers.Azure.Endpoint = v
	}
	if v, ok := root.get("providers", "azure", "api_key"); ok {
		cfg.Providers.Azure.APIKey = v
	}
	if v, ok := root.get("providers", "azure", "deployment"); ok {
		cfg.Providers.Azure.Deployment = v
	}
	if v, ok := root.get("execution", "auto_execute"); ok {
		cfg.Execution.AutoExecute = parseBool(v)
	}
	if v, ok := root.get("execution", "timeout"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Execution.Timeout = n
		}
	}
	if v, ok := root.get("execution", "shell_integration"); ok {
		cfg.Execution.ShellIntegration = parseBool(v)
	}
	if v, ok := root.get("safety", "level"); ok {
		cfg.Safety.Level = v
	}
	if v, ok := root.get("safety", "require_confirmation"); ok {
		cfg.Safety.RequireConfirmation = parseBool(v)
	}
	if v, ok := root.get("safety", "dry_run"); ok {
		cfg.Safety.DryRun = parseBool(v)
	}
	if v, ok := root.get("features", "explain"); ok {
		cfg.Features.Explain = parseBool(v)
	}
	if v, ok := root.get("features", "cache_commands"); ok {
		cfg.Features.CacheCommands = parseBool(v)
	}
	if v, ok := root.get("features", "learning_mode"); ok {
		cfg.Features.LearningMode = parseBool(v)
	}
	if v, ok := root.get("logging", "enabled"); ok {
		cfg.Logging.Enabled = parseBool(v)
	}
	if v, ok := root.get("logging", "level"); ok {
		cfg.Logging.Level = v
	}
}

func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "1":
		return true
	default:
		return false
	}
}

// Encode writes cfg as strict-subset YAML.
func Encode(cfg Config, w io.Writer) error {
	lines := []string{
		"provider: " + cfg.Provider,
		"model: " + cfg.Model,
		"",
		"providers:",
		"  ollama:",
		"    host: " + quoteIfNeeded(cfg.Providers.Ollama.Host),
		"    model: " + cfg.Providers.Ollama.Model,
		"  openai:",
		"    api_key: " + quoteIfNeeded(cfg.Providers.OpenAI.APIKey),
		"    model: " + cfg.Providers.OpenAI.Model,
		"  azure:",
		"    endpoint: " + quoteIfNeeded(cfg.Providers.Azure.Endpoint),
		"    api_key: " + quoteIfNeeded(cfg.Providers.Azure.APIKey),
		"    deployment: " + quoteIfNeeded(cfg.Providers.Azure.Deployment),
		"",
		"execution:",
		"  auto_execute: " + strconv.FormatBool(cfg.Execution.AutoExecute),
		"  timeout: " + strconv.Itoa(cfg.Execution.Timeout),
		"  shell_integration: " + strconv.FormatBool(cfg.Execution.ShellIntegration),
		"",
		"safety:",
		"  level: " + cfg.Safety.Level,
		"  require_confirmation: " + strconv.FormatBool(cfg.Safety.RequireConfirmation),
		"  dry_run: " + strconv.FormatBool(cfg.Safety.DryRun),
		"",
		"features:",
		"  explain: " + strconv.FormatBool(cfg.Features.Explain),
		"  cache_commands: " + strconv.FormatBool(cfg.Features.CacheCommands),
		"  learning_mode: " + strconv.FormatBool(cfg.Features.LearningMode),
		"",
		"logging:",
		"  enabled: " + strconv.FormatBool(cfg.Logging.Enabled),
		"  level: " + cfg.Logging.Level,
		"",
	}
	_, err := io.WriteString(w, strings.Join(lines, "\n"))
	return err
}

func quoteIfNeeded(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, ":#\n\"'") || strings.Contains(s, " ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}
