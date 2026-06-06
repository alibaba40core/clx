package pipeline

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/memory"
	"github.com/alibaba40core/clx/internal/parser"
)

// resolveChain tries resolvers in order. The first non-ErrNotFound result
// (success or hard error) wins. ErrNotFound propagates only if every
// resolver misses. Honors ctx between hops.
// aiHopIndex is the chain index of the AI resolver, or -1 when AI is not wired.
func resolveChain(ctx context.Context, req parser.Request, resolvers []intent.Resolver, logger *slog.Logger, aiHopIndex int, prog *progress) (intent.ResolvedIntent, error) {
	if len(resolvers) == 0 {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	var aiAttempted bool
	for i, r := range resolvers {
		if err := ctx.Err(); err != nil {
			return intent.ResolvedIntent{}, err
		}
		start := time.Now()
		var resolved intent.ResolvedIntent
		var err error
		if aiHopIndex >= 0 && i == aiHopIndex {
			stop := prog.Spin()
			resolved, err = r.Resolve(ctx, req)
			stop()
		} else {
			resolved, err = r.Resolve(ctx, req)
		}
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
		if aiHopIndex >= 0 && i == aiHopIndex {
			aiAttempted = true
		}
	}
	return intent.ResolvedIntent{}, &intent.MissError{AIAttempted: aiAttempted}
}

// buildResolvers assembles the resolution chain: memory, rules, optional cache, optional AI.
func buildResolvers(eng *intent.Engine, opts Options, cfg config.Config) []intent.Resolver {
	out := make([]intent.Resolver, 0, 4)
	if cfg.Memory.Enabled && opts.MemoryStore != nil {
		out = append(out, memory.NewResolver(opts.MemoryStore))
	}
	out = append(out, eng)
	if opts.Cache != nil {
		out = append(out, cache.AsResolver(opts.Cache, opts.Logger))
	}
	if opts.AIResolver != nil {
		ai := opts.AIResolver
		if opts.Cache != nil {
			ai = cache.WrapAIResolver(ai, opts.Cache, opts.Logger)
		}
		out = append(out, ai)
	}
	return out
}

// aiResolverIndex returns the resolver-chain index of the AI hop, or -1 when AI is not wired.
func aiResolverIndex(opts Options, cfg config.Config) int {
	if opts.AIResolver == nil {
		return -1
	}
	idx := 1
	if cfg.Memory.Enabled && opts.MemoryStore != nil {
		idx++
	}
	if opts.Cache != nil {
		idx++
	}
	return idx
}
