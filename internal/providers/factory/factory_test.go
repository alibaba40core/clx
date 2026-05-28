package factory

import (
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestNewFromConfigOllama(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	p, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "ollama" {
		t.Fatalf("name = %q", p.Name())
	}
}

func TestNewFromConfigOpenAINotImplemented(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Provider = "openai"
	_, err := NewFromConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewFromConfigOllamaMissingHost(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Providers.Ollama.Host = ""
	_, err := NewFromConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "ollama.host required") {
		t.Fatalf("err = %v", err)
	}
}
