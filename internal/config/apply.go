package config

import (
	"strconv"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

func applyNode(cfg *Config, root *yamlutil.Node) {
	if v, ok := root.GetString("provider"); ok {
		cfg.Provider = v
	}
	if v, ok := root.GetString("model"); ok {
		cfg.Model = v
	}
	if v, ok := root.GetString("providers", "primary"); ok {
		cfg.Providers.Primary = v
	}
	if v, ok := root.GetString("providers", "fallback"); ok {
		cfg.Providers.Fallback = v
	}
	if v, ok := root.GetString("providers", "timeout"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Providers.Timeout = n
		}
	}
	if root.Has("providers", "ollama", "host") {
		if v, ok := root.GetString("providers", "ollama", "host"); ok {
			cfg.Providers.Ollama.Host = v
		}
	}
	if root.Has("providers", "ollama", "model") {
		if v, ok := root.GetString("providers", "ollama", "model"); ok {
			cfg.Providers.Ollama.Model = v
		}
	}
	if v, ok := root.GetString("providers", "openai", "api_key"); ok {
		cfg.Providers.OpenAI.APIKey = v
	}
	if v, ok := root.GetString("providers", "openai", "model"); ok {
		cfg.Providers.OpenAI.Model = v
	}
	if v, ok := root.GetString("providers", "azure", "endpoint"); ok {
		cfg.Providers.Azure.Endpoint = v
	}
	if v, ok := root.GetString("providers", "azure", "api_key"); ok {
		cfg.Providers.Azure.APIKey = v
	}
	if v, ok := root.GetString("providers", "azure", "deployment"); ok {
		cfg.Providers.Azure.Deployment = v
	}
	if v, ok := root.GetString("providers", "gemini", "api_key"); ok {
		cfg.Providers.Gemini.APIKey = v
	}
	if v, ok := root.GetString("providers", "gemini", "model"); ok {
		cfg.Providers.Gemini.Model = v
	}
	if v, ok := root.GetString("execution", "auto_execute"); ok {
		cfg.Execution.AutoExecute = parseBool(v)
	}
	if v, ok := root.GetString("execution", "timeout"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Execution.Timeout = n
		}
	}
	if v, ok := root.GetString("execution", "shell_integration"); ok {
		cfg.Execution.ShellIntegration = parseBool(v)
	}
	if v, ok := root.GetString("safety", "mode"); ok {
		cfg.Safety.Mode = v
	} else if v, ok := root.GetString("safety", "level"); ok {
		// TODO(phase3): emit deprecation warning once Mode drives behavior.
		cfg.Safety.Mode = normalizeLegacyMode(v)
	}
	if v, ok := root.GetString("safety", "require_confirmation"); ok {
		cfg.Safety.RequireConfirmation = parseBool(v)
	}
	if v, ok := root.GetString("safety", "dry_run"); ok {
		cfg.Safety.DryRun = parseBool(v)
	}
	if v, ok := root.GetString("features", "explain"); ok {
		cfg.Features.Explain = parseBool(v)
	}
	if v, ok := root.GetString("features", "cache_commands"); ok {
		cfg.Features.CacheCommands = parseBool(v)
	}
	if v, ok := root.GetString("features", "learning_mode"); ok {
		cfg.Features.LearningMode = parseBool(v)
	}
	if v, ok := root.GetString("features", "ai_command_generation"); ok {
		cfg.Features.AICommandGeneration = parseBool(v)
	}
	if v, ok := root.GetString("cache", "max_entries"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Cache.MaxEntries = n
		}
	}
	if v, ok := root.GetString("cache", "ttl_days"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Cache.TTLDays = n
		}
	}
	if v, ok := root.GetString("cache", "max_disk_bytes"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Cache.MaxDiskBytes = n
		}
	}
	if v, ok := root.GetString("logging", "enabled"); ok {
		cfg.Logging.Enabled = parseBool(v)
	}
	if v, ok := root.GetString("logging", "level"); ok {
		cfg.Logging.Level = v
	}
}
