package config

import (
	"io"
	"strconv"
	"strings"
)

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
		"  mode: " + cfg.Safety.Mode,
		"  require_confirmation: " + strconv.FormatBool(cfg.Safety.RequireConfirmation),
		"  dry_run: " + strconv.FormatBool(cfg.Safety.DryRun),
		"",
		"features:",
		"  explain: " + strconv.FormatBool(cfg.Features.Explain),
		"  cache_commands: " + strconv.FormatBool(cfg.Features.CacheCommands),
		"  learning_mode: " + strconv.FormatBool(cfg.Features.LearningMode),
		"",
		"cache:",
		"  max_entries: " + strconv.Itoa(cfg.Cache.MaxEntries),
		"  ttl_days: " + strconv.Itoa(cfg.Cache.TTLDays),
		"  max_disk_bytes: " + strconv.Itoa(cfg.Cache.MaxDiskBytes),
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
