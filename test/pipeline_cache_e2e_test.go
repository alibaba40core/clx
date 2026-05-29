package test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/pipeline"
)

func newTestCacheStore(t *testing.T) *cache.Store {
	t.Helper()
	path, err := config.CacheIntentsPath()
	if err != nil {
		t.Fatal(err)
	}
	store, err := cache.Load(context.Background(), path, config.Default().Cache, nil)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestE2ECacheMissThenHitBypassesAI(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	store := newTestCacheStore(t)
	stub := &countingAIResolver{inner: stubAIResult{result: intent.ResolvedIntent{
		Intent: "current_dir", Params: map[string]string{}, Confidence: 0.9, Source: intent.SourceAI,
	}}}

	opts := pipeline.Options{
		Explain:    true,
		Cache:      store,
		AIResolver: stub,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}
	input := "totally unknown phrase xyz cache hit test"

	code, err := pipeline.Run(context.Background(), config.Default(), input, opts)
	if err != nil || code != 0 {
		t.Fatalf("first run code=%d err=%v", code, err)
	}
	if stub.calls.Load() != 1 {
		t.Fatalf("first run AI calls = %d, want 1", stub.calls.Load())
	}

	code, err = pipeline.Run(context.Background(), config.Default(), input, opts)
	if err != nil || code != 0 {
		t.Fatalf("second run code=%d err=%v", code, err)
	}
	if stub.calls.Load() != 1 {
		t.Fatalf("second run AI calls = %d, want 1 (cache hit)", stub.calls.Load())
	}
}

func TestE2ECacheRespectsFeatureFlag(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	stub := &countingAIResolver{inner: stubAIResult{result: intent.ResolvedIntent{
		Intent: "current_dir", Params: map[string]string{}, Confidence: 0.9, Source: intent.SourceAI,
	}}}

	cfg := config.Default()
	cfg.Features.CacheCommands = false
	opts := pipeline.Options{
		Explain:    true,
		Cache:      nil,
		AIResolver: stub,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}
	input := "unknown phrase feature flag cache test"

	for i := 0; i < 2; i++ {
		code, err := pipeline.Run(context.Background(), cfg, input, opts)
		if err != nil || code != 0 {
			t.Fatalf("run %d code=%d err=%v", i+1, code, err)
		}
	}
	if stub.calls.Load() != 2 {
		t.Fatalf("AI calls = %d, want 2 without cache resolver", stub.calls.Load())
	}
}

func TestE2ECacheProfileChangeMisses(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	store := newTestCacheStore(t)
	stub := &countingAIResolver{inner: stubAIResult{result: intent.ResolvedIntent{
		Intent: "current_dir", Params: map[string]string{}, Confidence: 0.9, Source: intent.SourceAI,
	}}}

	opts := pipeline.Options{
		Explain:    true,
		Cache:      store,
		AIResolver: stub,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}
	input := "unknown phrase profile change cache test"

	code, err := pipeline.Run(context.Background(), config.Default(), input, opts)
	if err != nil || code != 0 {
		t.Fatalf("first run code=%d err=%v", code, err)
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	p, err := environment.LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	profStore, err := environment.LoadStore(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	p.AvailableTools = []string{"cache-test-tool-only"}
	profStore.UpsertProfile(p)
	if err := environment.SaveStore(context.Background(), path, profStore); err != nil {
		t.Fatal(err)
	}

	code, err = pipeline.Run(context.Background(), config.Default(), input, opts)
	if err != nil || code != 0 {
		t.Fatalf("second run code=%d err=%v", code, err)
	}
	if stub.calls.Load() != 2 {
		t.Fatalf("AI calls = %d, want 2 after profile change", stub.calls.Load())
	}
}

func TestE2ECacheCorruptFileDegradesGracefully(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	path, err := config.CacheIntentsPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}

	store := newTestCacheStore(t)
	stub := &countingAIResolver{inner: stubAIResult{result: intent.ResolvedIntent{
		Intent: "current_dir", Params: map[string]string{}, Confidence: 0.9, Source: intent.SourceAI,
	}}}

	code, err := pipeline.Run(context.Background(), config.Default(), "unknown corrupt cache test", pipeline.Options{
		Explain:    true,
		Cache:      store,
		AIResolver: stub,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if stub.calls.Load() != 1 {
		t.Fatalf("AI calls = %d, want 1", stub.calls.Load())
	}
}
