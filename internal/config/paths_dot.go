package config

import (
	"fmt"
	"strconv"
	"strings"
)

var settablePaths = map[string]struct{}{
	"provider":                   {},
	"model":                      {},
	"providers.primary":          {},
	"providers.fallback":         {},
	"providers.timeout":          {},
	"providers.ollama.host":      {},
	"providers.ollama.model":     {},
	"providers.openai.api_key":   {},
	"providers.openai.model":     {},
	"providers.azure.endpoint":   {},
	"providers.azure.api_key":    {},
	"providers.azure.deployment": {},
	"providers.gemini.api_key":   {},
	"providers.gemini.model":     {},
}

// SetByPath mutates cfg at an allowlisted dot path.
func SetByPath(cfg *Config, path, value string) error {
	path = normalizeDotPath(path)
	if _, ok := settablePaths[path]; !ok {
		return fmt.Errorf("unknown or read-only config path %q", path)
	}
	value = strings.TrimSpace(value)
	switch path {
	case "provider":
		cfg.Provider = strings.ToLower(value)
	case "model":
		cfg.Model = value
	case "providers.primary":
		cfg.Providers.Primary = strings.ToLower(value)
	case "providers.fallback":
		cfg.Providers.Fallback = strings.ToLower(value)
	case "providers.timeout":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("providers.timeout must be an integer")
		}
		cfg.Providers.Timeout = n
	case "providers.ollama.host":
		cfg.Providers.Ollama.Host = value
	case "providers.ollama.model":
		cfg.Providers.Ollama.Model = value
	case "providers.openai.api_key":
		cfg.Providers.OpenAI.APIKey = value
	case "providers.openai.model":
		cfg.Providers.OpenAI.Model = value
	case "providers.azure.endpoint":
		cfg.Providers.Azure.Endpoint = value
	case "providers.azure.api_key":
		cfg.Providers.Azure.APIKey = value
	case "providers.azure.deployment":
		cfg.Providers.Azure.Deployment = value
	case "providers.gemini.api_key":
		cfg.Providers.Gemini.APIKey = value
	case "providers.gemini.model":
		cfg.Providers.Gemini.Model = value
	}
	return nil
}

// GetByPath returns the value at path. Secret paths are masked.
func GetByPath(cfg Config, path string) (string, error) {
	path = normalizeDotPath(path)
	if _, ok := settablePaths[path]; !ok {
		return "", fmt.Errorf("unknown config path %q", path)
	}
	var raw string
	switch path {
	case "provider":
		raw = cfg.Provider
	case "model":
		raw = cfg.Model
	case "providers.primary":
		raw = cfg.Providers.Primary
	case "providers.fallback":
		raw = cfg.Providers.Fallback
	case "providers.timeout":
		raw = strconv.Itoa(cfg.Providers.Timeout)
	case "providers.ollama.host":
		raw = cfg.Providers.Ollama.Host
	case "providers.ollama.model":
		raw = cfg.Providers.Ollama.Model
	case "providers.openai.api_key":
		raw = cfg.Providers.OpenAI.APIKey
	case "providers.openai.model":
		raw = cfg.Providers.OpenAI.Model
	case "providers.azure.endpoint":
		raw = cfg.Providers.Azure.Endpoint
	case "providers.azure.api_key":
		raw = cfg.Providers.Azure.APIKey
	case "providers.azure.deployment":
		raw = cfg.Providers.Azure.Deployment
	case "providers.gemini.api_key":
		raw = cfg.Providers.Gemini.APIKey
	case "providers.gemini.model":
		raw = cfg.Providers.Gemini.Model
	}
	if IsSecretPath(path) {
		return MaskSecret(raw), nil
	}
	return raw, nil
}

func normalizeDotPath(path string) string {
	return strings.ToLower(strings.TrimSpace(path))
}

// ProviderShowLines returns YAML-like lines for provider-related config (secrets masked).
func ProviderShowLines(cfg Config) []string {
	paths := []string{
		"provider",
		"model",
		"providers.primary",
		"providers.fallback",
		"providers.timeout",
		"providers.ollama.host",
		"providers.ollama.model",
		"providers.openai.api_key",
		"providers.openai.model",
		"providers.azure.endpoint",
		"providers.azure.api_key",
		"providers.azure.deployment",
		"providers.gemini.api_key",
		"providers.gemini.model",
	}
	lines := make([]string, 0, len(paths))
	for _, p := range paths {
		v, err := GetByPath(cfg, p)
		if err != nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", p, quoteDisplay(v)))
	}
	return lines
}

func quoteDisplay(s string) string {
	if s == `""` {
		return `""`
	}
	if strings.ContainsAny(s, ":#\n\"'") || strings.Contains(s, " ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

// SetProviderActive sets the active provider and clears providers.primary override.
func SetProviderActive(cfg *Config, name string) {
	cfg.Provider = strings.ToLower(strings.TrimSpace(name))
	cfg.Providers.Primary = ""
}
