package config

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

var (
	validProviders = map[string]struct{}{
		"ollama": {}, "openai": {}, "azure": {},
	}
	validSafetyLevels = map[string]struct{}{
		"safe": {}, "medium": {}, "full": {},
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

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks configuration invariants.
func Validate(c Config) error {
	if _, ok := validProviders[c.Provider]; !ok {
		return fmt.Errorf("invalid provider %q: must be ollama, openai, or azure", c.Provider)
	}
	if _, ok := validSafetyLevels[c.Safety.Level]; !ok {
		return fmt.Errorf("invalid safety.level %q: must be safe, medium, or full", c.Safety.Level)
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
	return nil
}
