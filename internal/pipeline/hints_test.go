package pipeline

import (
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestOllamaWSLHintLine(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Provider = "ollama"
	cfg.Providers.Ollama.Host = "http://localhost:11434"
	if got := OllamaWSLHintLine(cfg); got == "" {
		t.Fatal("expected hint for ollama localhost")
	}

	cfg.Provider = "openai"
	if got := OllamaWSLHintLine(cfg); got != "" {
		t.Fatalf("openai should not hint: %q", got)
	}

	cfg.Provider = "ollama"
	cfg.Providers.Ollama.Host = "http://192.168.1.10:11434"
	if got := OllamaWSLHintLine(cfg); got != "" {
		t.Fatalf("remote host should not hint: %q", got)
	}
}
