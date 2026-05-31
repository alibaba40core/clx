package providers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

const (
	defaultMinConfidence = 0.5
	// maxAdapterTimeout is the hard ceiling for AI resolution. Local Ollama on CPU
	// can exceed 30s on cold start or schema-constrained generation.
	maxAdapterTimeout = 180 * time.Second
)

// AdapterConfig tunes the Provider → intent.Resolver bridge.
type AdapterConfig struct {
	MinConfidence float64
	Timeout       time.Duration
}

// ErrorResolver returns a resolver that always fails with err (e.g. factory not-implemented).
func ErrorResolver(err error) intent.Resolver {
	if err == nil {
		return nil
	}
	return errorResolver{err: err}
}

type errorResolver struct {
	err error
}

func (e errorResolver) Resolve(context.Context, parser.Request) (intent.ResolvedIntent, error) {
	return intent.ResolvedIntent{}, e.err
}

// AsResolver wraps a Provider as an intent.Resolver for the pipeline chain.
func AsResolver(p Provider, eng *intent.Engine, logger *slog.Logger, cfg AdapterConfig) intent.Resolver {
	return &providerResolver{
		provider: p,
		engine:   eng,
		logger:   logger,
		cfg:      cfg,
	}
}

type providerResolver struct {
	provider Provider
	engine   *intent.Engine
	logger   *slog.Logger
	cfg      AdapterConfig
}

func (r *providerResolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	if err := ctx.Err(); err != nil {
		return intent.ResolvedIntent{}, err
	}

	profile, err := environment.LoadOrDetect(ctx)
	if err != nil {
		return intent.ResolvedIntent{}, err
	}

	intentReq := IntentRequest{
		RawInput:     req.RawInput,
		Profile:      profile,
		KnownIntents: r.engine.KnownIntents(),
		RuleParams:   ruleParamsFromEngine(r.engine),
		SkillHints:   SkillPromptsForEngine(r.engine),
	}

	timeout := r.cfg.Timeout
	if timeout <= 0 {
		timeout = maxAdapterTimeout
	}
	if timeout > maxAdapterTimeout {
		timeout = maxAdapterTimeout
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	resp, err := r.provider.ResolveIntent(callCtx, intentReq)
	latency := time.Since(start)

	if r.logger != nil {
		redacted := executor.Redact(req.RawInput)
		r.logger.Debug("ai resolve",
			"provider", r.provider.Name(),
			"latency_us", latency.Microseconds(),
			"input_len", len(redacted),
			"hit", err == nil,
		)
		if err == nil && resp != nil {
			r.logger.Debug("ai resolve result",
				"intent", resp.Intent,
				"confidence", resp.Confidence,
			)
		}
	}

	if err != nil {
		return intent.ResolvedIntent{}, mapAdapterError(err)
	}

	minConf := r.cfg.MinConfidence
	if minConf <= 0 {
		minConf = defaultMinConfidence
	}
	if resp.Confidence > 0 && resp.Confidence < minConf {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}

	return intent.ResolvedIntent{
		Intent:     resp.Intent,
		Params:     resp.Params,
		Confidence: resp.Confidence,
		Source:     intent.SourceAI,
	}, nil
}

func mapAdapterError(err error) error {
	switch {
	case errors.Is(err, ErrNoMatch), errors.Is(err, ErrInvalidResp):
		return intent.ErrNotFound
	case errors.Is(err, ErrRateLimited):
		return fmt.Errorf("provider rate limit exceeded: %w", err)
	case errors.Is(err, ErrAuth):
		return fmt.Errorf("provider authentication failed: %w", err)
	case errors.Is(err, ErrUnavailable):
		return fmt.Errorf("provider unavailable: %w", err)
	case errors.Is(err, ErrTimeout):
		return fmt.Errorf("provider timeout: %w", err)
	default:
		return err
	}
}

func ruleParamsFromEngine(eng *intent.Engine) map[string][]string {
	names := eng.KnownIntents()
	out := make(map[string][]string, len(names))
	for _, name := range names {
		rule, ok := eng.RuleForIntent(name)
		if !ok {
			continue
		}
		keys, err := intent.ExampleParamKeys(rule)
		if err != nil || len(keys) == 0 {
			continue
		}
		out[name] = keys
	}
	return out
}
