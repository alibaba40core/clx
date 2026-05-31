package pipeline

import (
	"fmt"
	"io"
	"strings"

	"github.com/alibaba40core/clx/internal/config"
)

// OllamaWSLHintLine returns a one-line hint when Ollama is configured on localhost
// and the caller should consider WSL host routing. Empty when not applicable.
func OllamaWSLHintLine(cfg config.Config) string {
	prov := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if cfg.Providers.Primary != "" {
		prov = strings.ToLower(strings.TrimSpace(cfg.Providers.Primary))
	}
	if prov != "ollama" {
		return ""
	}
	host := strings.TrimSpace(cfg.Providers.Ollama.Host)
	if host == "" {
		host = "http://localhost:11434"
	}
	if !ollamaHostIsLocalhost(host) {
		return ""
	}
	return "if Ollama runs in WSL, set providers.ollama.host to the WSL IP (see doc/provider-config.md)"
}

func ollamaHostIsLocalhost(host string) bool {
	lower := strings.ToLower(host)
	return strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1")
}

func printOllamaWSLHint(w io.Writer, cfg config.Config) {
	if line := OllamaWSLHintLine(cfg); line != "" {
		fmt.Fprintln(w, line)
	}
}
