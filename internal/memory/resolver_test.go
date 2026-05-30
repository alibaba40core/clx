package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

func TestResolverFollowUpAgain(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	ctx := context.Background()
	store, err := Open(ctx, "test-session", config.MemoryConfig{MaxEntriesPerSession: 8})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AppendCommand(ctx, CommandEntry{
		Intent: "search_text_in_file",
		Params: map[string]string{"pattern": "errors", "file": "logs.txt"},
	}); err != nil {
		t.Fatal(err)
	}

	r := NewResolver(store)
	got, err := r.Resolve(ctx, parser.Request{
		RawInput:  "again",
		Tokens:    []string{"again"},
		InputType: parser.InputNaturalLanguage,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "search_text_in_file" || got.Source != intent.SourceMemory {
		t.Fatalf("got %+v", got)
	}
	if got.Params["file"] != "logs.txt" {
		t.Fatalf("params=%v", got.Params)
	}
}

func TestResolverShortInputWithoutOverlapMisses(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	ctx := context.Background()
	store, err := Open(ctx, "sess", config.MemoryConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AppendCommand(ctx, CommandEntry{
		Intent: "current_dir",
		Params: map[string]string{},
	}); err != nil {
		t.Fatal(err)
	}
	_, err = NewResolver(store).Resolve(ctx, parser.Request{
		RawInput:  "unknown phrase xyz",
		Tokens:    []string{"unknown", "phrase", "xyz"},
		InputType: parser.InputNaturalLanguage,
	})
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestResolverMissWithoutHistory(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	ctx := context.Background()
	store, err := Open(ctx, "empty", config.MemoryConfig{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewResolver(store).Resolve(ctx, parser.Request{
		Tokens:    []string{"again"},
		InputType: parser.InputNaturalLanguage,
	})
	if !errors.Is(err, intent.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}
