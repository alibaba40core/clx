//go:build !lite

package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/alibaba40core/clx/internal/aliases"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/memory"
	"github.com/alibaba40core/clx/internal/parser"
)

func continueAfterFastMiss(ctx context.Context, cfg config.Config, opts *Options, eng *intent.Engine, req parser.Request, profile environment.SystemProfile, lazyAliases *aliases.LazyLookup, aliasLookup parser.AliasLookup) (int, error) {
	return runFullResolve(ctx, cfg, opts, eng, req, profile, lazyAliases, aliasLookup)
}

func runFullResolve(ctx context.Context, cfg config.Config, opts *Options, eng *intent.Engine, req parser.Request, profile environment.SystemProfile, lazyAliases *aliases.LazyLookup, aliasLookup parser.AliasLookup) (int, error) {
	fullProfile, err := environment.LoadProfile(ctx)
	if err != nil && !errors.Is(err, environment.ErrProfileNotFound) {
		fmt.Fprintf(opts.Stderr, "profile: %v\n", err)
		return 1, err
	}
	if err == nil {
		profile = fullProfile
	}

	if cfg.Memory.Enabled && opts.MemoryStore == nil {
		if store, merr := memory.Open(ctx, memory.DefaultSessionID(), cfg.Memory); merr == nil {
			opts.MemoryStore = store
		} else if opts.Logger != nil {
			opts.Logger.Warn("memory unavailable, continuing without session context", "err", merr)
		}
	}

	lc := newLazyCaches(cfg, opts.Logger)
	applyLazyCaches(opts, lc, ctx)
	lazyAIResolver := newLazyAI(cfg, eng, opts.Logger)
	if opts.AIResolver == nil && aiEnabled(cfg) {
		opts.AIResolver = lazyAIResolver
	}

	resolvers := buildResolvers(eng, *opts, cfg)
	resolved, err := resolveChain(ctx, req, resolvers, opts.Logger, aiResolverIndex(*opts, cfg), opts.prog)
	if err != nil {
		if isResolverMiss(err) {
			hintAliasMiss(*opts, aliasLookup, req)
			if opts.Provider == nil {
				opts.Provider = lazyAIResolver.Provider()
			}
			if code, handled, aiErr := tryAICommand(ctx, cfg, *opts, profile, req); handled {
				return code, aiErr
			}
		}
		return reportResolveError(cfg, *opts, req, err)
	}

	if resolved.Source != intent.SourceRule {
		if err := eng.ValidateResolved(resolved); err != nil {
			if opts.Logger != nil {
				opts.Logger.Warn("non-rule resolver output rejected",
					"source", resolved.Source.String(),
					"intent", resolved.Intent,
					"err", err,
				)
			}
			if ruleResolved, rerr := eng.Resolve(ctx, req); rerr == nil {
				resolved = ruleResolved
			} else if code, handled, aiErr := tryAICommand(ctx, cfg, *opts, profile, req); handled {
				return code, aiErr
			} else {
				fmt.Fprintf(opts.Stderr, "untrusted resolver output rejected: %v\n", err)
				return 1, err
			}
		}
	}

	translateProfile, err := profileForTranslate(ctx, eng, resolved, profile)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "profile: %v\n", err)
		return 1, err
	}

	if err := executor.ValidateIntentParams(resolved.Params); err != nil {
		fmt.Fprintf(opts.Stderr, "validation: %v\n", err)
		return 1, err
	}

	gen, err := generator.Translate(ctx, eng, resolved, translateProfile)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "translate: %v\n", err)
		return 1, err
	}

	_ = lazyAliases
	return executePlan(ctx, cfg, *opts, translateProfile, req, resolved, gen)
}
