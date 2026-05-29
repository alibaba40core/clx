// Package factory constructs Provider implementations from config without import cycles.
package factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/providers"
	"github.com/alibaba40core/clx/internal/providers/chain"
	"github.com/alibaba40core/clx/internal/providers/ollama"
	"github.com/alibaba40core/clx/internal/providers/openai"
)

// NewFromConfig constructs the configured Provider. No network I/O occurs here.
func NewFromConfig(cfg config.Config) (providers.Provider, error) {
	primaryName := config.EffectivePrimary(cfg)
	primary, err := newNamedProvider(cfg, primaryName)
	if err != nil {
		return nil, err
	}

	fallbackName := strings.ToLower(strings.TrimSpace(cfg.Providers.Fallback))
	if fallbackName == "" || fallbackName == primaryName {
		return primary, nil
	}

	fallback, err := newNamedProvider(cfg, fallbackName)
	if err != nil {
		return nil, err
	}
	return chain.New(primaryName, primary, fallbackName, fallback, nil), nil
}

func newNamedProvider(cfg config.Config, name string) (providers.Provider, error) {
	switch name {
	case "ollama":
		return newOllamaFromConfig(cfg)
	case "openai":
		return newOpenAIFromConfig(cfg)
	case "azure":
		return nil, fmt.Errorf("azure provider not implemented yet (Phase 2.3)")
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}

func newOllamaFromConfig(cfg config.Config) (providers.Provider, error) {
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
	return ollama.NewProvider(host, model, timeout)
}

func newOpenAIFromConfig(cfg config.Config) (providers.Provider, error) {
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
	return openai.NewProvider(apiKey, model, timeout)
}

func providerTimeout(cfg config.Config) time.Duration {
	timeout := time.Duration(cfg.Execution.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return timeout
}
