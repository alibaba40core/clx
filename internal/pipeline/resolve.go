package pipeline

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

// resolveChain tries resolvers in order. The first non-ErrNotFound result
// (success or hard error) wins. ErrNotFound propagates only if every
// resolver misses. Honors ctx between hops.
func resolveChain(ctx context.Context, req parser.Request, resolvers []intent.Resolver, logger *slog.Logger) (intent.ResolvedIntent, error) {
	if len(resolvers) == 0 {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	for i, r := range resolvers {
		if err := ctx.Err(); err != nil {
			return intent.ResolvedIntent{}, err
		}
		start := time.Now()
		resolved, err := r.Resolve(ctx, req)
		latency := time.Since(start)
		if logger != nil {
			logger.Debug("resolver hop",
				"index", i,
				"hit", err == nil,
				"latency_us", latency.Microseconds(),
			)
		}
		if err == nil {
			return resolved, nil
		}
		if !errors.Is(err, intent.ErrNotFound) {
			return intent.ResolvedIntent{}, err
		}
	}
	return intent.ResolvedIntent{}, intent.ErrNotFound
}

// buildResolvers assembles the resolution chain. Today: rules only.
// Phase 2.1 appends AI when opts.AIResolver != nil.
// Phase 2.2 inserts cache between rule and AI.
func buildResolvers(eng *intent.Engine, opts Options) []intent.Resolver {
	resolvers := []intent.Resolver{eng}
	if opts.AIResolver != nil {
		resolvers = append(resolvers, opts.AIResolver)
	}
	return resolvers
}
