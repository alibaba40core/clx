//go:build !lite

package pipeline

import (
	"context"
	"log/slog"
	"sync"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/providers"
	providerfactory "github.com/alibaba40core/clx/internal/providers/factory"
)

type lazyAI struct {
	cfg    config.Config
	eng    *intent.Engine
	logger *slog.Logger

	once     sync.Once
	provider providers.Provider
	resolver intent.Resolver
	initErr  error
}

func newLazyAI(cfg config.Config, eng *intent.Engine, logger *slog.Logger) *lazyAI {
	return &lazyAI{cfg: cfg, eng: eng, logger: logger}
}

func (l *lazyAI) resolverOnce() intent.Resolver {
	l.once.Do(func() {
		p, perr := providerfactory.NewFromConfig(l.cfg, l.logger)
		switch {
		case perr != nil:
			l.resolver = providers.ErrorResolver(perr)
			l.initErr = perr
		case p != nil:
			l.provider = p
			timeout := config.ProviderTimeout(l.cfg)
			l.resolver = providers.AsResolver(p, l.eng, l.logger, providers.AdapterConfig{
				Timeout: timeout,
			})
		}
	})
	return l.resolver
}

func (l *lazyAI) Provider() providers.Provider {
	l.resolverOnce()
	return l.provider
}

// Resolve implements intent.Resolver with lazy provider wiring.
func (l *lazyAI) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	r := l.resolverOnce()
	if r == nil {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	return r.Resolve(ctx, req)
}
