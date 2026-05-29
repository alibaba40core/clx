package chain

import (
	"context"
	"errors"
	"log/slog"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/providers"
)

// Provider tries primary first; on ErrUnavailable only, delegates to fallback (D9).
type Provider struct {
	primaryName  string
	primary      providers.Provider
	fallbackName string
	fallback     providers.Provider
	logger       *slog.Logger
}

// New returns a chain provider. Either child may be nil only if the caller guarantees
// it is never invoked; factory always passes non-nil providers.
func New(primaryName string, primary providers.Provider, fallbackName string, fallback providers.Provider, logger *slog.Logger) *Provider {
	return &Provider{
		primaryName:  primaryName,
		primary:      primary,
		fallbackName: fallbackName,
		fallback:     fallback,
		logger:       logger,
	}
}

// Name returns a composite identifier for logging.
func (c *Provider) Name() string {
	return c.primaryName + "+" + c.fallbackName
}

// ResolveIntent calls primary, then fallback only when primary returns ErrUnavailable.
func (c *Provider) ResolveIntent(ctx context.Context, req providers.IntentRequest) (*providers.IntentResponse, error) {
	resp, err := c.primary.ResolveIntent(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !errors.Is(err, providers.ErrUnavailable) {
		return nil, err
	}
	if c.logger != nil {
		c.logger.Debug("provider fallback",
			"primary", c.primaryName,
			"fallback", c.fallbackName,
		)
	}
	return c.fallback.ResolveIntent(ctx, req)
}

// Explain delegates to the primary provider.
func (c *Provider) Explain(ctx context.Context, gen generator.GeneratedCommand) (string, error) {
	return c.primary.Explain(ctx, gen)
}

var _ providers.Provider = (*Provider)(nil)
