package pipeline

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

type stubResolver struct {
	name   string
	result intent.ResolvedIntent
	err    error
	calls  int
}

func (s *stubResolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	s.calls++
	if err := ctx.Err(); err != nil {
		return intent.ResolvedIntent{}, err
	}
	if s.err != nil {
		return intent.ResolvedIntent{}, s.err
	}
	return s.result, nil
}

func TestResolveChainEmpty(t *testing.T) {
	_, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"pwd"}}, nil, nil, -1, nil)
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("got %v want ErrNotFound", err)
	}
}

func TestResolveChainSingleHit(t *testing.T) {
	r := &stubResolver{
		result: intent.ResolvedIntent{
			Intent: "current_dir",
			Source: intent.SourceRule,
		},
	}
	got, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"pwd"}}, []intent.Resolver{r}, nil, -1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "current_dir" || got.Source != intent.SourceRule {
		t.Fatalf("got %+v", got)
	}
	if r.calls != 1 {
		t.Fatalf("calls=%d want 1", r.calls)
	}
}

func TestResolveChainFirstMissSecondHit(t *testing.T) {
	miss := &stubResolver{err: intent.ErrNotFound}
	hit := &stubResolver{
		result: intent.ResolvedIntent{
			Intent: "find_file",
			Source: intent.SourceAI,
		},
	}
	got, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"locate", "x"}}, []intent.Resolver{miss, hit}, nil, -1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "find_file" || got.Source != intent.SourceAI {
		t.Fatalf("got %+v", got)
	}
	if miss.calls != 1 || hit.calls != 1 {
		t.Fatalf("miss=%d hit=%d", miss.calls, hit.calls)
	}
}

func TestResolveChainHardErrorPropagates(t *testing.T) {
	hard := errors.New("provider down")
	first := &stubResolver{err: hard}
	second := &stubResolver{
		result: intent.ResolvedIntent{Intent: "should_not_run"},
	}
	_, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"x"}}, []intent.Resolver{first, second}, nil, -1, nil)
	if !errors.Is(err, hard) {
		t.Fatalf("got %v want %v", err, hard)
	}
	if second.calls != 0 {
		t.Fatalf("second resolver called %d times", second.calls)
	}
}

func TestResolveChainCtxCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := resolveChain(ctx, parser.Request{Tokens: []string{"pwd"}}, []intent.Resolver{&stubResolver{}}, nil, -1, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveChainShortCircuitOnHit(t *testing.T) {
	first := &stubResolver{
		result: intent.ResolvedIntent{Intent: "a", Source: intent.SourceRule},
	}
	second := &stubResolver{
		result: intent.ResolvedIntent{Intent: "b", Source: intent.SourceAI},
	}
	got, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"pwd"}}, []intent.Resolver{first, second}, nil, -1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "a" || second.calls != 0 {
		t.Fatalf("got %+v second.calls=%d", got, second.calls)
	}
}

func TestBuildResolversRulesOnly(t *testing.T) {
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	resolvers := buildResolvers(eng, Options{}, cfg)
	if len(resolvers) != 1 {
		t.Fatalf("len=%d want 1", len(resolvers))
	}
}

func TestBuildResolversWithAI(t *testing.T) {
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	ai := &stubResolver{result: intent.ResolvedIntent{Intent: "x"}}
	opts := Options{AIResolver: ai}
	cfg := config.Default()
	resolvers := buildResolvers(eng, opts, cfg)
	if len(resolvers) != 2 {
		t.Fatalf("len=%d want 2", len(resolvers))
	}
	if aiResolverIndex(opts, cfg) != 1 {
		t.Fatalf("ai index = %d want 1", aiResolverIndex(opts, cfg))
	}
}

func TestResolveChainDebugLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	r := &stubResolver{
		result: intent.ResolvedIntent{Intent: "current_dir", Source: intent.SourceRule},
	}
	_, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"pwd"}}, []intent.Resolver{r}, logger, -1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected debug log output")
	}
	if !strings.Contains(buf.String(), "resolver hop") {
		t.Fatalf("log=%q", buf.String())
	}
}

func TestResolveChainAllMiss(t *testing.T) {
	miss1 := &stubResolver{err: intent.ErrNotFound}
	miss2 := &stubResolver{err: intent.ErrNotFound}
	_, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"unknown"}}, []intent.Resolver{miss1, miss2}, nil, 1, nil)
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	miss, ok := intent.AsMiss(err)
	if !ok || !miss.AIAttempted {
		t.Fatalf("want AIAttempted miss, got %v ok=%v", err, ok)
	}
}

func TestResolveChainAllMissNoAI(t *testing.T) {
	miss1 := &stubResolver{err: intent.ErrNotFound}
	_, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"unknown"}}, []intent.Resolver{miss1}, nil, -1, nil)
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	miss, ok := intent.AsMiss(err)
	if !ok || miss.AIAttempted {
		t.Fatalf("want rules-only miss, got %v", err)
	}
}

// Ensure resolveChain does not add measurable latency when logger is nil.
func TestResolveChainNoLogger(t *testing.T) {
	r := &stubResolver{
		result: intent.ResolvedIntent{Intent: "current_dir", Source: intent.SourceRule},
	}
	start := time.Now()
	_, err := resolveChain(context.Background(), parser.Request{Tokens: []string{"pwd"}}, []intent.Resolver{r}, nil, -1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("unexpected slow resolve")
	}
}
