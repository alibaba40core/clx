package config

import (
	"testing"
)

func TestSetByPathProviderFields(t *testing.T) {
	t.Parallel()
	cfg := Default()
	if err := SetByPath(&cfg, "providers.openai.model", "gpt-4.1"); err != nil {
		t.Fatal(err)
	}
	if cfg.Providers.OpenAI.Model != "gpt-4.1" {
		t.Fatalf("model = %q", cfg.Providers.OpenAI.Model)
	}
}

func TestSetByPathExecutionTimeout(t *testing.T) {
	t.Parallel()
	cfg := Default()
	if err := SetByPath(&cfg, "execution.timeout", "60"); err != nil {
		t.Fatal(err)
	}
	if cfg.Execution.Timeout != 60 {
		t.Fatalf("timeout = %d", cfg.Execution.Timeout)
	}
}

func TestSetByPathFeaturesCacheCommands(t *testing.T) {
	t.Parallel()
	cfg := Default()
	if err := SetByPath(&cfg, "features.cache_commands", "false"); err != nil {
		t.Fatal(err)
	}
	if cfg.Features.CacheCommands {
		t.Fatal("expected false")
	}
	got, err := GetByPath(cfg, "features.cache_commands")
	if err != nil || got != "false" {
		t.Fatalf("get = %q err=%v", got, err)
	}
}

func TestSetByPathRejectsUnknown(t *testing.T) {
	t.Parallel()
	cfg := Default()
	if err := SetByPath(&cfg, "safety.mode", "high"); err == nil {
		t.Fatal("expected error for non-allowlisted path")
	}
}

func TestGetByPathMasksSecrets(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.OpenAI.APIKey = "sk-abcdefghijklmnop"
	got, err := GetByPath(cfg, "providers.openai.api_key")
	if err != nil {
		t.Fatal(err)
	}
	if got == cfg.Providers.OpenAI.APIKey {
		t.Fatalf("secret not masked: %q", got)
	}
	if got != "****mnop" {
		t.Fatalf("got %q", got)
	}
}

func TestSetProviderActiveClearsPrimary(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.Primary = "ollama"
	SetProviderActive(&cfg, "openai")
	if cfg.Provider != "openai" {
		t.Fatalf("provider = %q", cfg.Provider)
	}
	if cfg.Providers.Primary != "" {
		t.Fatalf("primary should be cleared, got %q", cfg.Providers.Primary)
	}
}
