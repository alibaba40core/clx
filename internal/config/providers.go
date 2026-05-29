package config

import (
	"fmt"
	"strings"
)

// EffectivePrimary returns the active primary provider name (D7).
// providers.primary wins over top-level provider when set.
func EffectivePrimary(c Config) string {
	if p := strings.TrimSpace(c.Providers.Primary); p != "" {
		return strings.ToLower(p)
	}
	return strings.ToLower(strings.TrimSpace(c.Provider))
}

func normalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func validateProviderName(name, field string) error {
	if name == "" {
		return nil
	}
	if _, ok := validProviders[name]; !ok {
		return fmt.Errorf("invalid %s %q: must be ollama, openai, or azure", field, name)
	}
	return nil
}

func validateProviderSettings(c Config, name string) error {
	switch name {
	case "ollama":
		if strings.TrimSpace(c.Providers.Ollama.Host) == "" {
			return fmt.Errorf("providers.ollama.host required when %q is used", name)
		}
	case "openai":
		if strings.TrimSpace(c.Providers.OpenAI.APIKey) == "" {
			return fmt.Errorf("providers.openai.api_key required when %q is used", name)
		}
	case "azure":
		// Azure remains a factory stub in Phase 2.3; no extra field checks here.
	}
	return nil
}
