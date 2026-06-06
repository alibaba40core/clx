//go:build !lite

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/providers"
)

const aiCommandLabel = "ai-generated command"
const aiCommandMinConfidence = 0.5
const maxAICommandTimeout = 180 * time.Second

func tryAICommand(ctx context.Context, cfg config.Config, opts Options, profile environment.SystemProfile, req parser.Request) (code int, handled bool, err error) {
	if !cfg.Features.AICommandGeneration {
		return 0, false, nil
	}
	gen, ok := opts.Provider.(providers.CommandGenerator)
	if !ok || opts.Provider == nil {
		return 0, false, nil
	}

	timeout := config.ProviderTimeout(cfg)
	if timeout <= 0 || timeout > maxAICommandTimeout {
		timeout = maxAICommandTimeout
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	raw := req.RawInput
	if req.EffectiveInput != "" {
		raw = req.EffectiveInput
	}

	var resp *providers.CommandResponse
	var genErr error
	fromCache := false
	if opts.CommandCache != nil {
		key := cache.CommandKeyFor(raw, profile)
		if entry, ok := opts.CommandCache.Lookup(callCtx, key); ok {
			resp = cache.ToCommandResponse(entry)
			fromCache = true
		}
	}
	if resp == nil {
		stop := opts.prog.Spin()
		resp, genErr = gen.GenerateCommand(callCtx, providers.CommandRequest{
			RawInput: raw,
			Profile:  profile,
		})
		stop()
		if genErr != nil {
			return reportAICommandError(cfg, opts, genErr), true, genErr
		}
	}
	if resp == nil {
		fmt.Fprintf(opts.Stderr, "AI returned no command; try rephrasing the request\n")
		return 1, true, intent.ErrNotFound
	}

	if resp.Confidence > 0 && resp.Confidence < aiCommandMinConfidence {
		fmt.Fprintf(opts.Stderr, "AI was not confident enough to generate a command; try a more specific request\n")
		return 1, true, intent.ErrNotFound
	}

	shellHint := resp.Shell
	if shellHint == "" {
		shellHint = profile.Shell
	}

	gcmd, valErr := buildGeneratedFromAI(resp, shellHint, profile, raw)
	if valErr != nil && isAIValidationError(valErr) && !fromCache {
		stop := opts.prog.Spin()
		retryResp, retryErr := gen.GenerateCommand(callCtx, providers.CommandRequest{
			RawInput: raw,
			Profile:  profile,
			Feedback: valErr.Error(),
		})
		stop()
		if retryErr != nil {
			fmt.Fprintf(opts.Stderr, "AI command rejected as unsafe: %v\n", valErr)
			return 1, true, valErr
		}
		if retryResp != nil {
			resp = retryResp
			if resp.Shell != "" {
				shellHint = resp.Shell
			}
			gcmd, valErr = buildGeneratedFromAI(resp, shellHint, profile, raw)
		}
	}
	if valErr != nil {
		fmt.Fprintf(opts.Stderr, "AI command rejected as unsafe: %v\n", valErr)
		return 1, true, valErr
	}

	if !fromCache && opts.CommandCache != nil && resp != nil && !resp.HasChain() && !providers.ArgvHasChainConnector(resp.Argv) {
		key := cache.CommandKeyFor(raw, profile)
		if err := opts.CommandCache.Put(callCtx, key, resp); err != nil && opts.Logger != nil {
			opts.Logger.Warn("command cache write failed", "err", err)
		}
	}

	resolved := intent.ResolvedIntent{
		Intent:     aiCommandLabel,
		Confidence: resp.Confidence,
		Source:     intent.SourceAI,
	}

	code, err = executePlan(ctx, cfg, opts, profile, req, resolved, gcmd)
	return code, true, err
}

func isAIValidationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, executor.ErrAIArgvEmpty) ||
		errors.Is(err, executor.ErrAIArgvTooLong) ||
		errors.Is(err, executor.ErrAIArgvToken) ||
		errors.Is(err, executor.ErrChainEmpty) ||
		errors.Is(err, executor.ErrChainTooLong) ||
		errors.Is(err, executor.ErrChainExprCap) ||
		errors.Is(err, executor.ErrEmptyScriptArgv) ||
		errors.Is(err, executor.ErrScriptMetachar) ||
		errors.Is(err, executor.ErrCommandQuality)
}

func reportAICommandError(cfg config.Config, opts Options, err error) int {
	switch {
	case errors.Is(err, providers.ErrRateLimited):
		fmt.Fprintln(opts.Stderr, "AI provider rate limit exceeded (check quota/billing or try again later)")
	case errors.Is(err, providers.ErrAuth):
		fmt.Fprintln(opts.Stderr, "AI provider authentication failed (check API key in clx config show)")
	case errors.Is(err, providers.ErrInvalidResp):
		fmt.Fprintf(opts.Stderr, "AI returned an invalid command response (JSON/schema); try again or rephrase\n")
	case errors.Is(err, providers.ErrNoMatch):
		fmt.Fprintf(opts.Stderr, "AI could not generate a command for this request; try rephrasing\n")
	case errors.Is(err, providers.ErrUnavailable):
		fmt.Fprintf(opts.Stderr, "AI provider unavailable: %v\n", err)
		printOllamaWSLHint(opts.Stderr, cfg)
	case errors.Is(err, providers.ErrTimeout):
		fmt.Fprintf(opts.Stderr, "AI provider timed out generating a command\n")
	default:
		fmt.Fprintf(opts.Stderr, "AI command generation failed: %v\n", err)
	}
	return 1
}
