package pipeline

import (
	"io"
	"log/slog"
	"os"

	"github.com/alibaba40core/clx/internal/intent"
)

// Options configures pipeline execution.
type Options struct {
	Explain bool
	DryRun  bool
	Yes     bool
	Logger  *slog.Logger
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer

	// AIResolver, when non-nil, runs after the rule engine misses.
	// Phase 2.1 wires the Ollama/OpenAI provider here.
	AIResolver intent.Resolver
}

// WithDefaults fills nil writers and stdin.
func (o *Options) WithDefaults() {
	if o.Stdin == nil {
		o.Stdin = os.Stdin
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
}
