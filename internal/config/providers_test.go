package config

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

func TestEffectivePrimaryPrefersProvidersPrimary(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Provider = "ollama"
	cfg.Providers.Primary = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	if got := EffectivePrimary(cfg); got != "openai" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyPrimaryFallback(t *testing.T) {
	t.Parallel()
	root, err := yamlutil.Decode(strings.NewReader(`
providers:
  primary: ollama
  fallback: openai
  openai:
    api_key: sk-test
`))
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	applyNode(&cfg, root)
	if cfg.Providers.Primary != "ollama" || cfg.Providers.Fallback != "openai" {
		t.Fatalf("cfg = %+v", cfg.Providers)
	}
	if cfg.Providers.OpenAI.APIKey != "sk-test" {
		t.Fatalf("api_key = %q", cfg.Providers.OpenAI.APIKey)
	}
}

func TestValidateFallbackRequiresOpenAIKey(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.Fallback = "openai"
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "api_key required") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateFallbackMustDifferFromPrimary(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.Primary = "ollama"
	cfg.Providers.Fallback = "ollama"
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateOllamaHostURL(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.Ollama.Host = "not-a-url"
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "http") {
		t.Fatalf("err = %v", err)
	}
}

func TestProviderTimeoutPrefersProvidersTimeout(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Execution.Timeout = 30
	cfg.Providers.Timeout = 45
	if got := ProviderTimeout(cfg); got != 45*time.Second {
		t.Fatalf("got %v", got)
	}
}

func TestValidateProvidersTimeoutNegative(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.Timeout = -1
	if err := Validate(cfg); err == nil {
		t.Fatal("expected error")
	}
}

func TestEncodePrimaryFallbackRoundTrip(t *testing.T) {
	t.Parallel()
	cfg := Default()
	cfg.Providers.Primary = "ollama"
	cfg.Providers.Fallback = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	var buf bytes.Buffer
	if err := Encode(cfg, &buf); err != nil {
		t.Fatal(err)
	}
	root, err := yamlutil.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := Default()
	applyNode(&got, root)
	if got.Providers.Primary != "ollama" || got.Providers.Fallback != "openai" {
		t.Fatalf("got = %+v", got.Providers)
	}
}
