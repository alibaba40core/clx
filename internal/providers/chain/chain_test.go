package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/providers"
)

type stubProvider struct {
	name   string
	resp   *providers.IntentResponse
	err    error
	calls  int
}

func (s *stubProvider) Name() string { return s.name }

func (s *stubProvider) ResolveIntent(context.Context, providers.IntentRequest) (*providers.IntentResponse, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func (s *stubProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return "", nil
}

func TestChainPrimarySuccess(t *testing.T) {
	t.Parallel()
	primary := &stubProvider{name: "ollama", resp: &providers.IntentResponse{Intent: "list_dir", Confidence: 0.9}}
	fallback := &stubProvider{name: "openai", resp: &providers.IntentResponse{Intent: "current_dir", Confidence: 0.9}}
	c := New("ollama", primary, "openai", fallback, nil)
	resp, err := c.ResolveIntent(context.Background(), providers.IntentRequest{})
	if err != nil || resp.Intent != "list_dir" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
	if fallback.calls != 0 {
		t.Fatalf("fallback calls = %d", fallback.calls)
	}
}

func TestChainFallbackOnUnavailable(t *testing.T) {
	t.Parallel()
	primary := &stubProvider{name: "ollama", err: providers.ErrUnavailable}
	fallback := &stubProvider{name: "openai", resp: &providers.IntentResponse{Intent: "current_dir", Confidence: 0.9}}
	c := New("ollama", primary, "openai", fallback, nil)
	resp, err := c.ResolveIntent(context.Background(), providers.IntentRequest{})
	if err != nil || resp.Intent != "current_dir" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
	if fallback.calls != 1 {
		t.Fatalf("fallback calls = %d", fallback.calls)
	}
}

func TestChainNoFallbackOnTimeout(t *testing.T) {
	t.Parallel()
	primary := &stubProvider{name: "ollama", err: providers.ErrTimeout}
	fallback := &stubProvider{name: "openai", resp: &providers.IntentResponse{Intent: "current_dir", Confidence: 0.9}}
	c := New("ollama", primary, "openai", fallback, nil)
	_, err := c.ResolveIntent(context.Background(), providers.IntentRequest{})
	if !errors.Is(err, providers.ErrTimeout) {
		t.Fatalf("err = %v", err)
	}
	if fallback.calls != 0 {
		t.Fatalf("fallback calls = %d", fallback.calls)
	}
}

func TestChainBothUnavailable(t *testing.T) {
	t.Parallel()
	primary := &stubProvider{name: "ollama", err: providers.ErrUnavailable}
	fallback := &stubProvider{name: "openai", err: providers.ErrUnavailable}
	c := New("ollama", primary, "openai", fallback, nil)
	_, err := c.ResolveIntent(context.Background(), providers.IntentRequest{})
	if !errors.Is(err, providers.ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
}
