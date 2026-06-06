package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/intent"
)

type lazyCaches struct {
	cfg    config.Config
	logger *slog.Logger

	once     sync.Once
	intents  *cache.Store
	explain  *cache.ExplainStore
	commands *cache.CommandStore
}

func newLazyCaches(cfg config.Config, logger *slog.Logger) *lazyCaches {
	return &lazyCaches{cfg: cfg, logger: logger}
}

func (lc *lazyCaches) load(ctx context.Context) {
	lc.once.Do(func() {
		if !lc.cfg.Features.CacheCommands || lc.cfg.Cache.MaxEntries <= 0 {
			return
		}
		if path, err := config.CacheIntentsPath(); err == nil {
			if store, lerr := cache.Load(ctx, path, lc.cfg.Cache, lc.logger); lerr == nil {
				lc.intents = store
			} else if lc.logger != nil {
				lc.logger.Warn("cache unavailable", "err", lerr)
			}
		}
		if path, err := config.CacheExplanationsPath(); err == nil {
			if store, lerr := cache.LoadExplain(ctx, path, lc.cfg.Cache, lc.logger); lerr == nil {
				lc.explain = store
			} else if lc.logger != nil {
				lc.logger.Warn("explain cache unavailable", "err", lerr)
			}
		}
		if path, err := config.CacheCommandsPath(); err == nil {
			if store, lerr := cache.LoadCommands(ctx, path, lc.cfg.Cache, lc.logger); lerr == nil {
				lc.commands = store
			} else if lc.logger != nil {
				lc.logger.Warn("command cache unavailable", "err", lerr)
			}
		}
	})
}

func (lc *lazyCaches) Intents(ctx context.Context) *cache.Store {
	lc.load(ctx)
	return lc.intents
}

func (lc *lazyCaches) Explain(ctx context.Context) *cache.ExplainStore {
	lc.load(ctx)
	return lc.explain
}

func (lc *lazyCaches) Commands(ctx context.Context) *cache.CommandStore {
	lc.load(ctx)
	return lc.commands
}

func applyLazyCaches(opts *Options, lc *lazyCaches, ctx context.Context) {
	if opts.Cache == nil {
		opts.Cache = lc.Intents(ctx)
	}
	if opts.ExplainCache == nil {
		opts.ExplainCache = lc.Explain(ctx)
	}
	if opts.CommandCache == nil {
		opts.CommandCache = lc.Commands(ctx)
	}
}

func ensureEngine(ctx context.Context, opts *Options) (*intent.Engine, error) {
	if opts.Engine != nil {
		return opts.Engine, nil
	}
	eng, err := intent.NewEngineWithOverlay(ctx, opts.Logger)
	if err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}
	opts.Engine = eng
	return eng, nil
}
