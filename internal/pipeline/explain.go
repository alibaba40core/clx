package pipeline

import (
	"context"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/providers"
)

const explainTimeout = 2 * time.Second

func shouldEnrichExplanation(opts Options, resolved intent.ResolvedIntent) bool {
	if !opts.Explain {
		return false
	}
	if resolved.Source != intent.SourceAI && resolved.Source != intent.SourceCache {
		return false
	}
	return opts.Provider != nil
}

func enrichExplanation(ctx context.Context, opts Options, resolved intent.ResolvedIntent, gen generator.GeneratedCommand) string {
	static := gen.Explanation
	key := cache.ExplainKeyFor(resolved.Intent, gen)

	if opts.ExplainCache != nil {
		if entry, ok := opts.ExplainCache.Lookup(ctx, key); ok {
			return entry.Text
		}
	}

	callCtx, cancel := context.WithTimeout(ctx, explainTimeout)
	defer cancel()

	text, err := opts.Provider.Explain(callCtx, gen)
	if err != nil || strings.TrimSpace(text) == "" {
		if opts.Logger != nil && err != nil {
			opts.Logger.Debug("ai explain fallback to static",
				"provider", opts.Provider.Name(),
				"err", providers.RedactError(err),
			)
		}
		return static
	}
	text = strings.TrimSpace(text)

	if opts.ExplainCache != nil {
		if err := opts.ExplainCache.Put(ctx, key, text); err != nil && opts.Logger != nil {
			opts.Logger.Warn("explain cache write failed", "err", err)
		}
	}
	return text
}
