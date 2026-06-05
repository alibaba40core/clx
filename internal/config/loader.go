package config

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

var (
	// "none" disables the AI provider entirely (rules-only mode). It is a valid
	// provider value so users are never forced to run Ollama or any LLM.
	validProviders = map[string]struct{}{
		"none": {}, "ollama": {}, "openai": {}, "azure": {}, "gemini": {},
	}
	validSafetyModes = map[string]struct{}{
		"low": {}, "medium": {}, "high": {}, "custom": {},
	}
	validLogLevels = map[string]struct{}{
		"debug": {}, "info": {}, "warn": {}, "error": {},
	}
)

// Load reads and validates configuration from path.
func Load(ctx context.Context, path string) (Config, error) {
	if err := ctx.Err(); err != nil {
		return Config{}, err
	}

	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, err
	}
	defer f.Close()

	if err := ctx.Err(); err != nil {
		return Config{}, err
	}

	root, err := yamlutil.Decode(f)
	if err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	applyNode(&cfg, root)

	if err := DecryptConfig(&cfg); err != nil {
		return Config{}, err
	}

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks configuration invariants.
func Validate(c Config) error {
	if _, ok := validProviders[c.Provider]; !ok {
		return fmt.Errorf("invalid provider %q: must be none, ollama, openai, azure, or gemini", c.Provider)
	}
	if _, ok := validSafetyModes[c.Safety.Mode]; !ok {
		return fmt.Errorf("invalid safety.mode %q: must be low, medium, high, or custom", c.Safety.Mode)
	}
	if c.Execution.Timeout <= 0 {
		return fmt.Errorf("execution.timeout must be > 0")
	}
	level := c.Logging.Level
	if level == "" {
		level = "info"
	}
	if _, ok := validLogLevels[level]; !ok {
		return fmt.Errorf("invalid logging.level %q", c.Logging.Level)
	}
	if c.Cache.MaxEntries <= 0 {
		return fmt.Errorf("cache.max_entries must be > 0")
	}
	if c.Cache.TTLDays <= 0 {
		return fmt.Errorf("cache.ttl_days must be > 0")
	}
	if c.Cache.MaxDiskBytes <= 0 {
		return fmt.Errorf("cache.max_disk_bytes must be > 0")
	}
	if c.Providers.Timeout < 0 {
		return fmt.Errorf("providers.timeout must be >= 0")
	}

	primary := EffectivePrimary(c)
	if err := validateProviderName(primary, "providers.primary"); err != nil {
		return err
	}
	if err := validateProviderSettings(c, primary); err != nil {
		return err
	}

	fallback := normalizeProviderName(c.Providers.Fallback)
	if fallback != "" {
		if err := validateProviderName(fallback, "providers.fallback"); err != nil {
			return err
		}
		if fallback == primary {
			return fmt.Errorf("providers.fallback must differ from effective primary %q", primary)
		}
		if err := validateProviderSettings(c, fallback); err != nil {
			return err
		}
	}

	return nil
}
