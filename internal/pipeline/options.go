package pipeline

import (
	"io"
	"log/slog"
	"os"

	"github.com/alibaba40core/clx/internal/aliases"
	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/memory"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/providers"
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

	// Engine, when set, is used for rule resolution and ValidateResolved.
	// If nil, Run builds an overlay engine internally (tests).
	Engine *intent.Engine

	// AIResolver, when non-nil, runs after the rule engine misses.
	// Phase 2.1 wires the Ollama/OpenAI provider here.
	AIResolver intent.Resolver

	// Cache, when non-nil, sits between rules and AI in the resolver chain.
	Cache *cache.Store

	// Provider, when non-nil, supplies AI Explain for --explain on AI/Cache hits.
	Provider providers.Provider

	// ExplainCache, when non-nil, caches AI explanations at ~/.clx/cache/explanations.json.
	ExplainCache *cache.ExplainStore

	// CommandCache, when non-nil, caches AI command-generation at ~/.clx/cache/commands.json.
	CommandCache *cache.CommandStore

	// AliasStore, when set, is used for parser-stage alias expansion (tests).
	AliasStore *aliases.Store

	// MemoryStore, when set, is used for session follow-ups and append (tests).
	MemoryStore *memory.Store
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
