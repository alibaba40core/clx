// Package factory constructs Provider implementations from config without import cycles.
package factory

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/providers"
	"github.com/alibaba40core/clx/internal/providers/ollama"
)

// NewFromConfig constructs the configured Provider. No network I/O occurs here.
func NewFromConfig(cfg config.Config) (providers.Provider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "ollama":
		return newOllamaFromConfig(cfg)
	case "openai":
		return nil, fmt.Errorf("openai provider not implemented yet (Phase 2.3)")
	case "azure":
		return nil, fmt.Errorf("azure provider not implemented yet (Phase 2.3)")
	default:
		return nil, fmt.Errorf("unknown provider %q", cfg.Provider)
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
	timeout := time.Duration(cfg.Execution.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return ollama.NewProvider(host, model, timeout)
}
