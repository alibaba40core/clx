package ollama

import (
	"context"
	"errors"
	"time"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/providers"
)

// Provider implements providers.Provider via a local Ollama HTTP client.
type Provider struct {
	client *Client
}

// NewProvider wires host, model, and HTTP timeout. Does not contact Ollama until ResolveIntent.
func NewProvider(host, model string, timeout time.Duration) (*Provider, error) {
	client, err := NewClient(host, model, timeout)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client}, nil
}

// Name returns the provider identifier for logging and config.
func (p *Provider) Name() string {
	return "ollama"
}

// ResolveIntent builds the global prompt, schema, and calls Ollama /api/chat.
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
	case errors.Is(err, errInvalidResp):
		return providers.ErrInvalidResp
	case errors.Is(err, errNoMatch):
		return providers.ErrNoMatch
	default:
		return err
	}
}

// Explain is deferred to Phase 2.4.
func (p *Provider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return "", nil
}
