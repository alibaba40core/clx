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

func TestSetByPathRejectsUnknown(t *testing.T) {
	t.Parallel()
	cfg := Default()
	if err := SetByPath(&cfg, "execution.timeout", "60"); err == nil {
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
