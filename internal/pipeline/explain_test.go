package pipeline

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/policy"
	"github.com/alibaba40core/clx/internal/providers"
)

func explainTestHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()
	p, err := environment.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(p)
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

type explainStubProvider struct {
	text string
	err  error
	calls int
}

func (s *explainStubProvider) Name() string { return "stub" }

func (s *explainStubProvider) ResolveIntent(context.Context, providers.IntentRequest) (*providers.IntentResponse, error) {
	return nil, providers.ErrNoMatch
}

func (s *explainStubProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	s.calls++
	return s.text, s.err
}

func TestRunExplainAIEnrichment(t *testing.T) {
	explainTestHome(t)
	stub := &explainStubProvider{text: "AI explains listing files."}
	var stdout bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "unknown phrase", Options{
		Explain: true,
		Provider: stub,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "list_dir", Params: map[string]string{}, Source: intent.SourceAI, Confidence: 0.9,
		}},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if stub.calls != 1 {
		t.Fatalf("explain calls = %d", stub.calls)
	}
	if !strings.Contains(stdout.String(), "AI explains listing files.") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunExplainRulePathNoAIExplain(t *testing.T) {
	explainTestHome(t)
	stub := &explainStubProvider{text: "should not appear"}
	var stdout bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "git status", Options{
		Explain:  true,
		Provider: stub,
		Stdout:   &stdout,
		Stderr:   &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if stub.calls != 0 {
		t.Fatalf("explain calls = %d, want 0 for rule hit", stub.calls)
	}
}

func TestRunExplainAIFallbackStatic(t *testing.T) {
	explainTestHome(t)
	stub := &explainStubProvider{err: providers.ErrUnavailable}
	var stdout bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "unknown", Options{
		Explain: true,
		Provider: stub,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "list_dir", Source: intent.SourceAI, Confidence: 0.9,
		}},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stdout.String(), "Explanation:") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if strings.Contains(stdout.String(), "should not appear") {
		t.Fatal("unexpected AI text")
	}
}

func TestExplainFallbackRedactsErrorInDebugLog(t *testing.T) {
	explainTestHome(t)
	var logBuf strings.Builder
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	secret := "sk-secret1234567890abcdef"
	stub := &explainStubProvider{err: errString("provider unavailable: " + secret)}
	var stdout bytes.Buffer
	_, err := Run(context.Background(), config.Default(), "unknown", Options{
		Explain: true,
		Logger:  logger,
		Provider: stub,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "list_dir", Source: intent.SourceAI, Confidence: 0.9,
		}},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(logBuf.String(), secret) {
		t.Fatalf("secret in logs: %s", logBuf.String())
	}
}

type errString string

func (e errString) Error() string { return string(e) }

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
