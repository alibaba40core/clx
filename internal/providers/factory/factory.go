// Package factory constructs Provider implementations from config without import cycles.
package factory

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/providers"
	"github.com/alibaba40core/clx/internal/providers/chain"
	"github.com/alibaba40core/clx/internal/providers/gemini"
	"github.com/alibaba40core/clx/internal/providers/ollama"
	"github.com/alibaba40core/clx/internal/providers/openai"
)

// NewFromConfig constructs the configured Provider. No network I/O occurs here.
// A provider of "none" (or empty) returns (nil, nil): rules-only mode with no AI
// wired, so CLX never requires Ollama or any other LLM to run.
func NewFromConfig(cfg config.Config, logger *slog.Logger) (providers.Provider, error) {
	primaryName := config.EffectivePrimary(cfg)
	if primaryName == "" || primaryName == "none" {
		return nil, nil
	}
	primary, err := newNamedProvider(cfg, primaryName, logger)
	if err != nil {
		return nil, err
	}

	fallbackName := strings.ToLower(strings.TrimSpace(cfg.Providers.Fallback))
	if fallbackName == "" || fallbackName == primaryName {
		return primary, nil
	}

	fallback, err := newNamedProvider(cfg, fallbackName, logger)
	if err != nil {
		return nil, err
	}
	return chain.New(primaryName, primary, fallbackName, fallback, logger), nil
}

func newNamedProvider(cfg config.Config, name string, logger *slog.Logger) (providers.Provider, error) {
	switch name {
	case "none", "":
		return nil, nil
	case "ollama":
		return newOllamaFromConfig(cfg, logger)
	case "openai":
		return newOpenAIFromConfig(cfg, logger)
	case "gemini":
		return newGeminiFromConfig(cfg, logger)
	case "azure":
		return nil, fmt.Errorf("azure provider not implemented yet (Phase 2.3)")
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}

func newOllamaFromConfig(cfg config.Config, logger *slog.Logger) (providers.Provider, error) {
	host := strings.TrimSpace(cfg.Providers.Ollama.Host)
	if host == "" {
		return nil, fmt.Errorf("ollama.host required when provider is ollama")
	}
	model := strings.TrimSpace(cfg.Providers.Ollama.Model)
	if model == "" {
		model = strings.TrimSpace(cfg.Model)
	}
	if model == "" {
		model = config.DefaultOllamaModel
	}
	timeout := providerTimeout(cfg)
	return ollama.NewProvider(host, model, timeout, logger)
}

func newOpenAIFromConfig(cfg config.Config, logger *slog.Logger) (providers.Provider, error) {
	apiKey := strings.TrimSpace(cfg.Providers.OpenAI.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("openai.api_key required when provider is openai")
	}
	model := strings.TrimSpace(cfg.Providers.OpenAI.Model)
	if model == "" {
		model = strings.TrimSpace(cfg.Model)
	}
	if model == "" {
		model = "gpt-4.1-mini"
	}
	timeout := providerTimeout(cfg)
	return openai.NewProvider(apiKey, model, timeout, logger)
}

func providerTimeout(cfg config.Config) time.Duration {
	return config.ProviderTimeout(cfg)
}

func newGeminiFromConfig(cfg config.Config, logger *slog.Logger) (providers.Provider, error) {
	apiKey := strings.TrimSpace(cfg.Providers.Gemini.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("gemini.api_key required when provider is gemini")
	}
	model := strings.TrimSpace(cfg.Providers.Gemini.Model)
	if model == "" {
		model = strings.TrimSpace(cfg.Model)
	}
	if model == "" {
		model = config.DefaultGeminiModel
	}
	timeout := providerTimeout(cfg)
	return gemini.NewProvider(apiKey, model, timeout, logger)
}
