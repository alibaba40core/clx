package test

import (
	"bytes"
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/pipeline"
	"github.com/alibaba40core/clx/internal/providers"
	providerfactory "github.com/alibaba40core/clx/internal/providers/factory"
)

// countingAIResolver records how many times the AI path was invoked.
type countingAIResolver struct {
	calls atomic.Int32
	inner intent.Resolver
}

func (c *countingAIResolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	c.calls.Add(1)
	if c.inner != nil {
		return c.inner.Resolve(ctx, req)
	}
	return intent.ResolvedIntent{}, intent.ErrNotFound
}

type stubAIResult struct {
	result intent.ResolvedIntent
	err    error
}

func (s stubAIResult) Resolve(context.Context, parser.Request) (intent.ResolvedIntent, error) {
	if s.err != nil {
		return intent.ResolvedIntent{}, s.err
	}
	return s.result, nil
}

func TestE2EAIRulePathBypassesAI(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	var stderr bytes.Buffer
	stub := &countingAIResolver{}
	code, err := pipeline.Run(context.Background(), config.Default(), "git status", pipeline.Options{
		Explain:    true,
		AIResolver: stub,
		Stdout:     &bytes.Buffer{},
		Stderr:     &stderr,
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	if stub.calls.Load() != 0 {
		t.Fatalf("AI resolver calls = %d, want 0", stub.calls.Load())
	}
}

func TestE2EAIRulesMissThenAIHit(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	var stdout bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "totally unknown phrase xyz", pipeline.Options{
		Explain: true,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "current_dir", Params: map[string]string{}, Confidence: 0.9,
			Source: intent.SourceAI,
		}},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stdout.String(), "current_dir") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestE2EAIRejectsMaliciousIntent(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	var stderr bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "unknown xyz", pipeline.Options{
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "rm_rf_slash", Params: map[string]string{}, Source: intent.SourceAI,
		}},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "untrusted resolver output rejected") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestE2EAIProviderDownHardFails(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	var stderr bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "find all files modified today", pipeline.Options{
		AIResolver: providers.AsResolver(
			stubProvider{err: providers.ErrUnavailable},
			mustEngine(t), nil, providers.AdapterConfig{}),
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "provider unavailable") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestE2EAIProviderFlagOverrideOpenAI(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Provider = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	p, err := providerfactory.NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatalf("factory err = %v", err)
	}
	if p.Name() != "openai" {
		t.Fatalf("name = %q", p.Name())
	}
}

func TestE2EAIProviderFlagDisablesFallback(t *testing.T) {
	cfg := config.Default()
	cfg.Providers.Fallback = "openai"
	cfg.Providers.OpenAI.APIKey = "sk-test"
	cfg.Provider = "ollama"
	cfg.Providers.Fallback = ""
	p, err := providerfactory.NewFromConfig(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "ollama" {
		t.Fatalf("name = %q, want single ollama provider", p.Name())
	}
}

func TestE2EAILowConfidenceTreatedAsMiss(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	var stderr bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "find all files modified today", pipeline.Options{
		AIResolver: providers.AsResolver(stubProvider{resp: &providers.IntentResponse{
			Intent: "list_dir", Params: map[string]string{}, Confidence: 0.2,
		}}, mustEngine(t), nil, providers.AdapterConfig{MinConfidence: 0.5}),
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "AI could not map") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func mustEngine(t *testing.T) *intent.Engine {
	t.Helper()
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

// stubProvider in test package mirrors providers.stubProvider for adapter wiring.
type stubProvider struct {
	resp *providers.IntentResponse
	err  error
}

func (stubProvider) Name() string { return "stub" }

func (s stubProvider) ResolveIntent(context.Context, providers.IntentRequest) (*providers.IntentResponse, error) {
	return s.resp, s.err
}

func (stubProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return "", nil
}

var _ providers.Provider = stubProvider{}
