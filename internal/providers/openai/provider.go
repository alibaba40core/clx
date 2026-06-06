package openai

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/providers"
)

// Provider implements providers.Provider via the OpenAI chat completions API.
type Provider struct {
	client *Client
}

// NewProvider wires apiKey, model, and HTTP timeout. Does not contact OpenAI until ResolveIntent.
func NewProvider(apiKey, model string, timeout time.Duration, logger *slog.Logger) (*Provider, error) {
	client, err := NewClient(apiKey, model, "", timeout, logger)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client}, nil
}

// Name returns the provider identifier for logging and config.
func (p *Provider) Name() string {
	return "openai"
}

// ResolveIntent builds the global prompt, schema, and calls OpenAI chat completions.
func (p *Provider) ResolveIntent(ctx context.Context, req providers.IntentRequest) (*providers.IntentResponse, error) {
	system, user, err := providers.BuildPrompt(req)
	if err != nil {
		return nil, err
	}
	schema, err := providers.BuildResponseSchema(req)
	if err != nil {
		return nil, err
	}
	out, err := p.client.Chat(ctx, system, user, schema)
	if err != nil {
		return nil, mapClientError(err)
	}
	return &providers.IntentResponse{
		Intent:     out.Intent,
		Params:     out.Params,
		Confidence: out.Confidence,
	}, nil
}

func mapClientError(err error) error {
	switch {
	case errors.Is(err, errUnavailable):
		return providers.ErrUnavailable
	case errors.Is(err, errTimeout):
		return providers.ErrTimeout
	case errors.Is(err, errInvalidResp), errors.Is(err, providers.ErrInvalidResp):
		return providers.ErrInvalidResp
	case errors.Is(err, providers.ErrRateLimited):
		return providers.ErrRateLimited
	case errors.Is(err, providers.ErrAuth):
		return providers.ErrAuth
	case errors.Is(err, errNoMatch):
		return providers.ErrNoMatch
	default:
		return err
	}
}

// Explain calls OpenAI for a plain-text command explanation (Phase 2.4).
func (p *Provider) Explain(ctx context.Context, gen generator.GeneratedCommand) (string, error) {
	profile, err := environment.ProfileForResolver(ctx)
	if err != nil {
		return "", err
	}
	system, user, err := providers.BuildExplainPrompt(gen, profile)
	if err != nil {
		return "", err
	}
	text, err := p.client.ExplainChat(ctx, system, user)
	if err != nil {
		return "", mapClientError(err)
	}
	return text, nil
}
