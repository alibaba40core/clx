package providers

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
)

func TestSentinelErrorIdentity(t *testing.T) {
	t.Parallel()
	sentinels := []error{
		ErrUnavailable,
		ErrTimeout,
		ErrInvalidResp,
		ErrNoMatch,
		ErrRateLimited,
		ErrAuth,
	}
	for i, want := range sentinels {
		for j, other := range sentinels {
			if i == j {
				if !errors.Is(want, other) {
					t.Fatalf("errors.Is(%v, %v) = false, want true", want, other)
				}
				wrapped := fmt.Errorf("wrap: %w", want)
				if !errors.Is(wrapped, want) {
					t.Fatalf("errors.Is(wrapped, %v) = false", want)
				}
				continue
			}
			if errors.Is(want, other) {
				t.Fatalf("sentinel %v should not match %v", want, other)
			}
		}
	}
}

type nullProvider struct{}

func (nullProvider) Name() string { return "null" }

func (nullProvider) ResolveIntent(context.Context, IntentRequest) (*IntentResponse, error) {
	return nil, nil
}

func (nullProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return "", nil
}

func TestProviderInterfaceShape(t *testing.T) {
	t.Parallel()
	var _ Provider = nullProvider{}
	var _ Provider = (*nullProvider)(nil)
}
