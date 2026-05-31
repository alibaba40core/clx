package config

import (
	"fmt"
	"strconv"
	"strings"
)

var settablePaths = map[string]struct{}{
	"provider":                         {},
	"model":                            {},
	"providers.primary":                {},
	"providers.fallback":             {},
	"providers.timeout":                {},
	"providers.ollama.host":            {},
	"providers.ollama.model":           {},
	"providers.openai.api_key":       {},
	"providers.openai.model":         {},
	"providers.azure.endpoint":       {},
	"providers.azure.api_key":        {},
	"providers.azure.deployment":     {},
	"providers.gemini.api_key":       {},
	"providers.gemini.model":         {},
	"features.explain":               {},
	"features.cache_commands":        {},
	"features.ai_command_generation": {},
	"features.learning_mode":         {},
	"cache.max_entries":              {},
	"cache.ttl_days":                 {},
	"cache.max_disk_bytes":           {},
	"memory.enabled":                   {},
	"memory.max_entries_per_session": {},
	"memory.max_sessions":            {},
	"memory.ttl_days":                {},
	"execution.auto_execute":         {},
	"execution.timeout":              {},
	"execution.shell_integration":    {},
	"logging.enabled":                  {},
	"logging.level":                    {},
	"aliases.max_aliases":            {},
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
	case "features.explain":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("features.explain must be true or false")
		}
		cfg.Features.Explain = b
	case "features.cache_commands":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("features.cache_commands must be true or false")
		}
		cfg.Features.CacheCommands = b
	case "features.ai_command_generation":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("features.ai_command_generation must be true or false")
		}
		cfg.Features.AICommandGeneration = b
	case "features.learning_mode":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("features.learning_mode must be true or false")
		}
		cfg.Features.LearningMode = b
	case "cache.max_entries":
		n, err := parsePositiveInt(value, "cache.max_entries")
		if err != nil {
			return err
		}
		cfg.Cache.MaxEntries = n
	case "cache.ttl_days":
		n, err := parsePositiveInt(value, "cache.ttl_days")
		if err != nil {
			return err
		}
		cfg.Cache.TTLDays = n
	case "cache.max_disk_bytes":
		n, err := parsePositiveInt(value, "cache.max_disk_bytes")
		if err != nil {
			return err
		}
		cfg.Cache.MaxDiskBytes = n
	case "memory.enabled":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("memory.enabled must be true or false")
		}
		cfg.Memory.Enabled = b
	case "memory.max_entries_per_session":
		n, err := parsePositiveInt(value, "memory.max_entries_per_session")
		if err != nil {
			return err
		}
		cfg.Memory.MaxEntriesPerSession = n
	case "memory.max_sessions":
		n, err := parsePositiveInt(value, "memory.max_sessions")
		if err != nil {
			return err
		}
		cfg.Memory.MaxSessions = n
	case "memory.ttl_days":
		n, err := parsePositiveInt(value, "memory.ttl_days")
		if err != nil {
			return err
		}
		cfg.Memory.TTLDays = n
	case "execution.auto_execute":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("execution.auto_execute must be true or false")
		}
		cfg.Execution.AutoExecute = b
	case "execution.timeout":
		n, err := parsePositiveInt(value, "execution.timeout")
		if err != nil {
			return err
		}
		cfg.Execution.Timeout = n
	case "execution.shell_integration":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("execution.shell_integration must be true or false")
		}
		cfg.Execution.ShellIntegration = b
	case "logging.enabled":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("logging.enabled must be true or false")
		}
		cfg.Logging.Enabled = b
	case "logging.level":
		level := strings.ToLower(value)
		if _, ok := validLogLevels[level]; !ok {
			return fmt.Errorf("logging.level must be debug, info, warn, or error")
		}
		cfg.Logging.Level = level
	case "aliases.max_aliases":
		n, err := parsePositiveInt(value, "aliases.max_aliases")
		if err != nil {
			return err
		}
		cfg.Aliases.MaxAliases = n
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
	case "features.explain":
		raw = strconv.FormatBool(cfg.Features.Explain)
	case "features.cache_commands":
		raw = strconv.FormatBool(cfg.Features.CacheCommands)
	case "features.ai_command_generation":
		raw = strconv.FormatBool(cfg.Features.AICommandGeneration)
	case "features.learning_mode":
		raw = strconv.FormatBool(cfg.Features.LearningMode)
	case "cache.max_entries":
		raw = strconv.Itoa(cfg.Cache.MaxEntries)
	case "cache.ttl_days":
		raw = strconv.Itoa(cfg.Cache.TTLDays)
	case "cache.max_disk_bytes":
		raw = strconv.Itoa(cfg.Cache.MaxDiskBytes)
	case "memory.enabled":
		raw = strconv.FormatBool(cfg.Memory.Enabled)
	case "memory.max_entries_per_session":
		raw = strconv.Itoa(cfg.Memory.MaxEntriesPerSession)
	case "memory.max_sessions":
		raw = strconv.Itoa(cfg.Memory.MaxSessions)
	case "memory.ttl_days":
		raw = strconv.Itoa(cfg.Memory.TTLDays)
	case "execution.auto_execute":
		raw = strconv.FormatBool(cfg.Execution.AutoExecute)
	case "execution.timeout":
		raw = strconv.Itoa(cfg.Execution.Timeout)
	case "execution.shell_integration":
		raw = strconv.FormatBool(cfg.Execution.ShellIntegration)
	case "logging.enabled":
		raw = strconv.FormatBool(cfg.Logging.Enabled)
	case "logging.level":
		raw = cfg.Logging.Level
	case "aliases.max_aliases":
		raw = strconv.Itoa(cfg.Aliases.MaxAliases)
	}
	if IsSecretPath(path) {
		return MaskSecret(raw), nil
	}
	return raw, nil
}

func parsePositiveInt(value, field string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", field)
	}
	return n, nil
}

func normalizeDotPath(path string) string {
	return strings.ToLower(strings.TrimSpace(path))
}

// ProviderShowLines returns YAML-like lines for provider-related config (secrets masked).
func ProviderShowLines(cfg Config) []string {
	lines := make([]string, 0, len(providerShowPaths()))
	for _, p := range providerShowPaths() {
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
