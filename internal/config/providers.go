package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// EffectivePrimary returns the active primary provider name (D7).
// providers.primary wins over top-level provider when set.
func EffectivePrimary(c Config) string {
	if p := strings.TrimSpace(c.Providers.Primary); p != "" {
		return strings.ToLower(p)
	}
	return strings.ToLower(strings.TrimSpace(c.Provider))
}

// ProviderTimeout returns the AI provider HTTP/resolver timeout.
// providers.timeout wins when > 0; else execution.timeout; else 30s.
func ProviderTimeout(c Config) time.Duration {
	if c.Providers.Timeout > 0 {
		return time.Duration(c.Providers.Timeout) * time.Second
	}
	if c.Execution.Timeout > 0 {
		return time.Duration(c.Execution.Timeout) * time.Second
	}
	return 30 * time.Second
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
		host := strings.TrimSpace(c.Providers.Ollama.Host)
		if host == "" {
			return fmt.Errorf("providers.ollama.host required when %q is used", name)
		}
		if err := validateHTTPURL(host, "providers.ollama.host"); err != nil {
			return err
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

func validateHTTPURL(raw, field string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: invalid URL: %w", field, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s: must use http or https scheme", field)
	}
	if u.Host == "" {
		return fmt.Errorf("%s: must include a host", field)
	}
	return nil
}
