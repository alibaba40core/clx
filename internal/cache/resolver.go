package cache

import (
	"context"
	"log/slog"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

// AsResolver returns a read-only intent.Resolver backed by the cache store.
func AsResolver(s *Store, logger *slog.Logger) intent.Resolver {
	if s == nil {
		return nil
	}
	return &readResolver{store: s, logger: logger}
}

// WrapAIResolver wraps an AI resolver with write-through caching on AI hits.
func WrapAIResolver(ai intent.Resolver, s *Store, logger *slog.Logger) intent.Resolver {
	if ai == nil {
		return nil
	}
	if s == nil {
		return ai
	}
	return &writeResolver{inner: ai, store: s, logger: logger}
}

type readResolver struct {
	store  *Store
	logger *slog.Logger
}

func (r *readResolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	if err := ctx.Err(); err != nil {
		return intent.ResolvedIntent{}, err
	}
	profile, err := environment.LoadOrDetect(ctx)
	if err != nil {
		return intent.ResolvedIntent{}, err
	}
	key := KeyFor(req, profile)
	entry, ok := r.store.Lookup(ctx, key)
	if !ok {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	return intent.ResolvedIntent{
		Intent:     entry.Intent,
		Params:     cloneParams(entry.Params),
		Confidence: entry.Confidence,
		Source:     intent.SourceCache,
	}, nil
}

type writeResolver struct {
	inner  intent.Resolver
	store  *Store
	logger *slog.Logger
}

func (r *writeResolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	resolved, err := r.inner.Resolve(ctx, req)
	if err != nil {
		return intent.ResolvedIntent{}, err
	}
	if resolved.Source != intent.SourceAI {
		return resolved, nil
	}

	profile, perr := environment.LoadOrDetect(ctx)
	if perr != nil {
		if r.logger != nil {
			r.logger.Warn("cache write skipped: profile unavailable", "err", perr)
		}
		return resolved, nil
	}

	key := KeyFor(req, profile)
	if err := r.store.Put(ctx, key, resolved.Intent, resolved.Params, resolved.Confidence); err != nil {
		if r.logger != nil {
			r.logger.Warn("cache write failed", "err", err)
		}
	}
	return resolved, nil
}
