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
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/providers"
)

// aiCommandLabel is shown in the Intent field for AI-generated commands.
const aiCommandLabel = "ai-generated command"

// aiCommandMinConfidence drops AI commands the model is not reasonably sure about.
const aiCommandMinConfidence = 0.5

// maxAICommandTimeout caps AI command generation latency.
const maxAICommandTimeout = 180 * time.Second

// tryAICommand attempts the hybrid AI command-generation fallback.
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
	if opts.CommandCache != nil {
		key := cache.CommandKeyFor(raw, profile)
		if entry, ok := opts.CommandCache.Lookup(callCtx, key); ok {
			resp = cache.ToCommandResponse(entry)
		}
	}
	if resp == nil {
		resp, genErr = gen.GenerateCommand(callCtx, providers.CommandRequest{
			RawInput: raw,
			Profile:  profile,
		})
		if genErr != nil {
			return reportAICommandError(cfg, opts, genErr), true, genErr
		}
		if resp != nil && opts.CommandCache != nil {
			key := cache.CommandKeyFor(raw, profile)
			if err := opts.CommandCache.Put(callCtx, key, resp); err != nil && opts.Logger != nil {
				opts.Logger.Warn("command cache write failed", "err", err)
			}
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
	if vErr := executor.ValidateGeneratedArgv(resp.Argv, shellHint); vErr != nil {
		fmt.Fprintf(opts.Stderr, "AI command rejected as unsafe: %v\n", vErr)
		return 1, true, vErr
	}
	gcmd := generator.NewAICommand(resp.Argv, resp.Shell, resp.Explanation, profile)

	resolved := intent.ResolvedIntent{
		Intent:     aiCommandLabel,
		Confidence: resp.Confidence,
		Source:     intent.SourceAI,
	}

	code, err = executePlan(ctx, cfg, opts, profile, req, resolved, gcmd)
	return code, true, err
}

// reportAICommandError prints a user-facing message for a provider-side failure
// during command generation and returns the pipeline exit code.
func reportAICommandError(cfg config.Config, opts Options, err error) int {
	switch {
	case errors.Is(err, providers.ErrRateLimited):
		fmt.Fprintln(opts.Stderr, "AI provider rate limit exceeded (check quota/billing or try again later)")
	case errors.Is(err, providers.ErrAuth):
		fmt.Fprintln(opts.Stderr, "AI provider authentication failed (check API key in clx config show)")
	case errors.Is(err, providers.ErrNoMatch), errors.Is(err, providers.ErrInvalidResp):
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
