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
	if v, ok := root.GetString("safety", "level"); ok {
		cfg.Safety.Level = v
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
	if v, ok := root.GetString("logging", "enabled"); ok {
		cfg.Logging.Enabled = parseBool(v)
	}
	if v, ok := root.GetString("logging", "level"); ok {
		cfg.Logging.Level = v
	}
}
