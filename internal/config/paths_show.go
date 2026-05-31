package config

import "fmt"

// ConfigShowLines returns a sectioned summary of runtime config (secrets masked).
func ConfigShowLines(cfg Config) []string {
	sections := []struct {
		title string
		paths []string
	}{
		{"# providers", providerShowPaths()},
		{"# features", []string{
			"features.explain",
			"features.cache_commands",
			"features.ai_command_generation",
			"features.learning_mode",
		}},
		{"# cache", []string{
			"cache.max_entries",
			"cache.ttl_days",
			"cache.max_disk_bytes",
		}},
		{"# memory", []string{
			"memory.enabled",
			"memory.max_entries_per_session",
			"memory.max_sessions",
			"memory.ttl_days",
		}},
		{"# execution", []string{
			"execution.auto_execute",
			"execution.timeout",
			"execution.shell_integration",
		}},
		{"# logging", []string{
			"logging.enabled",
			"logging.level",
		}},
		{"# aliases", []string{
			"aliases.max_aliases",
		}},
	}
	var lines []string
	for _, sec := range sections {
		lines = append(lines, sec.title)
		for _, p := range sec.paths {
			v, err := GetByPath(cfg, p)
			if err != nil {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: %s", p, quoteDisplay(v)))
		}
		lines = append(lines, "")
	}
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func providerShowPaths() []string {
	return []string{
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
}
