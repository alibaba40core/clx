package factory

import (
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/providers/chain"
)

func TestNewFromConfigOllama(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	p, err := NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "ollama" {
		t.Fatalf("name = %q", p.Name())
	}
}

func TestNewFromConfigOpenAI(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Provider = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	p, err := NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openai" {
		t.Fatalf("name = %q", p.Name())
	}
}

func TestNewFromConfigOpenAIMissingKey(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Provider = "openai"
	_, err := NewFromConfig(cfg, nil)
	if err == nil || !strings.Contains(err.Error(), "api_key required") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewFromConfigOpenAINotImplementedRemoved(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Provider = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	p, err := NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("expected provider")
	}
}

func TestNewFromConfigOllamaMissingHost(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Providers.Ollama.Host = ""
	_, err := NewFromConfig(cfg, nil)
	if err == nil || !strings.Contains(err.Error(), "ollama.host required") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewFromConfigChain(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Providers.Fallback = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	p, err := NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "ollama+openai" {
		t.Fatalf("name = %q", p.Name())
	}
	if _, ok := p.(*chain.Provider); !ok {
		t.Fatalf("type = %T", p)
	}
}

func TestNewFromConfigNoChainWithoutFallback(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	p, err := NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.(*chain.Provider); ok {
		t.Fatal("expected single provider without fallback")
	}
}

func TestNewFromConfigUsesProvidersPrimary(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Providers.Primary = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	p, err := NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openai" {
		t.Fatalf("name = %q", p.Name())
	}
}
