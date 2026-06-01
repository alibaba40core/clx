package cache

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

type stubResolver struct {
	result intent.ResolvedIntent
	err    error
	calls  int
}

func (s *stubResolver) Resolve(context.Context, parser.Request) (intent.ResolvedIntent, error) {
	s.calls++
	if s.err != nil {
		return intent.ResolvedIntent{}, s.err
	}
	return s.result, nil
}

func seedProfile(t *testing.T, profile environment.SystemProfile) {
	t.Helper()
	t.Setenv("CLX_HOME", t.TempDir())
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(profile)
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

func TestAsResolverHitAndMiss(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	profile, err := environment.LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(dir, "cache", "intents.json")
	s := testStore(t, cachePath, config.Default().Cache)
	req := parser.Request{InputType: parser.InputNaturalLanguage, Tokens: []string{"show", "files"}}
	key := KeyFor(req, profile)
	if err := s.Put(context.Background(), key, "list_dir", map[string]string{}, 0.9); err != nil {
		t.Fatal(err)
	}

	r := AsResolver(s, nil)
	got, err := r.Resolve(context.Background(), req)
	if err != nil || got.Intent != "list_dir" || got.Source != intent.SourceCache {
		t.Fatalf("got=%+v err=%v", got, err)
	}

	_, err = r.Resolve(context.Background(), parser.Request{InputType: parser.InputNaturalLanguage, Tokens: []string{"other"}})
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestWrapAIResolverWriteThrough(t *testing.T) {
	seedProfile(t, environment.SystemProfile{OS: "linux", Shell: "bash"})
	cachePath := filepath.Join(t.TempDir(), "intents.json")
	s := testStore(t, cachePath, config.Default().Cache)

	inner := &stubResolver{result: intent.ResolvedIntent{
		Intent: "current_dir", Params: map[string]string{}, Confidence: 0.9, Source: intent.SourceAI,
	}}
	wrapped := WrapAIResolver(inner, s, nil)
	req := parser.Request{InputType: parser.InputNaturalLanguage, Tokens: []string{"where", "am", "i"}}

	resolved, err := wrapped.Resolve(context.Background(), req)
	if err != nil || resolved.Intent != "current_dir" {
		t.Fatalf("resolved=%+v err=%v", resolved, err)
	}
	if inner.calls != 1 {
		t.Fatalf("inner calls = %d", inner.calls)
	}

	read := AsResolver(s, nil)
	got, err := read.Resolve(context.Background(), req)
	if err != nil || got.Intent != "current_dir" {
		t.Fatalf("cache hit=%+v err=%v", got, err)
	}
}

func TestWrapAIResolverSkipsNonAI(t *testing.T) {
	seedProfile(t, environment.SystemProfile{OS: "linux", Shell: "bash"})
	cachePath := filepath.Join(t.TempDir(), "intents.json")
	s := testStore(t, cachePath, config.Default().Cache)
	inner := &stubResolver{result: intent.ResolvedIntent{
		Intent: "list_dir", Source: intent.SourceRule,
	}}
	wrapped := WrapAIResolver(inner, s, nil)
	req := parser.Request{InputType: parser.InputNaturalLanguage, Tokens: []string{"ls"}}
	_, err := wrapped.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	read := AsResolver(s, nil)
	_, err = read.Resolve(context.Background(), req)
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("expected cache miss, err=%v", err)
	}
}

func TestWrapAIResolverSkipsOnError(t *testing.T) {
	seedProfile(t, environment.SystemProfile{OS: "linux", Shell: "bash"})
	s := testStore(t, filepath.Join(t.TempDir(), "intents.json"), config.Default().Cache)
	inner := &stubResolver{err: intent.ErrNotFound}
	wrapped := WrapAIResolver(inner, s, nil)
	_, err := wrapped.Resolve(context.Background(), parser.Request{Tokens: []string{"x"}})
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestAsResolverProfileChangeMisses(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	// The read resolver computes its cache key from the profile returned by
	// environment.LoadOrDetect, which keys on the runner's *actual* detected
	// OS/shell. Build the persisted profile from that detected base so the
	// resolver actually reads it, then vary a non-store-key field
	// (AvailableTools) to model a profile refresh: the store key stays the
	// same but the cache key changes, so a prior entry must miss.
	detected, err := environment.LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	path, _ := config.SystemProfilePath()

	p1 := detected
	p1.AvailableTools = []string{"git"}
	storeProf := environment.NewProfileStore()
	storeProf.UpsertProfile(p1)
	if err := environment.SaveStore(context.Background(), path, storeProf); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(dir, "cache", "intents.json")
	s := testStore(t, cachePath, config.Default().Cache)
	req := parser.Request{InputType: parser.InputNaturalLanguage, Tokens: []string{"pwd"}}
	if err := s.Put(context.Background(), KeyFor(req, p1), "current_dir", nil, 0.9); err != nil {
		t.Fatal(err)
	}

	p2 := p1
	p2.AvailableTools = []string{"git", "docker"}
	storeProf.UpsertProfile(p2)
	if err := environment.SaveStore(context.Background(), path, storeProf); err != nil {
		t.Fatal(err)
	}

	r := AsResolver(s, nil)
	_, err = r.Resolve(context.Background(), req)
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("expected miss after profile change, err=%v", err)
	}
}
