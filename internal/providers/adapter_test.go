package providers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

type stubProvider struct {
	resp *IntentResponse
	err  error
}

func (stubProvider) Name() string { return "stub" }

func (s stubProvider) ResolveIntent(context.Context, IntentRequest) (*IntentResponse, error) {
	return s.resp, s.err
}

func (stubProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return "", nil
}

var _ Provider = stubProvider{}

func TestAdapterMapsUnavailable(t *testing.T) {
	t.Parallel()
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	r := AsResolver(stubProvider{err: ErrUnavailable}, eng, nil, AdapterConfig{})
	_, err = r.Resolve(context.Background(), parser.Request{RawInput: "find logs"})
	if err == nil || !strings.Contains(err.Error(), "provider unavailable") {
		t.Fatalf("err = %v", err)
	}
}

func TestAdapterMapsNoMatchToErrNotFound(t *testing.T) {
	t.Parallel()
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	r := AsResolver(stubProvider{err: ErrNoMatch}, eng, nil, AdapterConfig{})
	_, err = r.Resolve(context.Background(), parser.Request{RawInput: "x"})
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestAdapterLowConfidenceIsMiss(t *testing.T) {
	t.Parallel()
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	r := AsResolver(stubProvider{resp: &IntentResponse{
		Intent: "list_dir", Params: map[string]string{}, Confidence: 0.1,
	}}, eng, nil, AdapterConfig{MinConfidence: 0.5})
	_, err = r.Resolve(context.Background(), parser.Request{RawInput: "x"})
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestAdapterRedactsInDebugLogs(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	r := AsResolver(stubProvider{resp: &IntentResponse{
		Intent: "list_dir", Params: map[string]string{}, Confidence: 0.9,
	}}, eng, logger, AdapterConfig{})
	_, _ = r.Resolve(context.Background(), parser.Request{
		RawInput: "deploy api_key=secret123",
	})
	if strings.Contains(buf.String(), "secret123") {
		t.Fatalf("secret in logs: %s", buf.String())
	}
}
