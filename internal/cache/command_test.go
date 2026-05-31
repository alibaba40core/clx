package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/providers"
)

func sampleProfileForCommand() environment.SystemProfile {
	return environment.SystemProfile{
		OS:             "linux",
		Shell:          "bash",
		AvailableTools: []string{"git", "rg"},
	}
}

func testCommandStore(t *testing.T, path string, cfg config.CacheConfig) *CommandStore {
	t.Helper()
	s, err := LoadCommands(context.Background(), path, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC) }
	return s
}

func TestCommandKeyForDeterministic(t *testing.T) {
	t.Parallel()
	raw := "totally unknown phrase"
	p := sampleProfileForCommand()
	k1 := CommandKeyFor(raw, p)
	k2 := CommandKeyFor(raw, p)
	if k1 != k2 {
		t.Fatalf("non-deterministic: %q vs %q", k1, k2)
	}
	p2 := p
	p2.Shell = "zsh"
	if k1 == CommandKeyFor(raw, p2) {
		t.Fatal("profile change must change key")
	}
}

func TestCommandCachePutLookupRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "commands.json")
	s := testCommandStore(t, path, config.Default().Cache)
	ctx := context.Background()
	profile := sampleProfileForCommand()
	raw := "list all pdf files"
	key := CommandKeyFor(raw, profile)

	resp := &providers.CommandResponse{
		Argv:        []string{"find", ".", "-name", "*.pdf"},
		Shell:       "bash",
		Explanation: "find PDFs",
		Confidence:  0.9,
	}
	if err := s.Put(ctx, key, resp); err != nil {
		t.Fatal(err)
	}
	entry, ok := s.Lookup(ctx, key)
	if !ok || len(entry.Argv) != 4 || entry.Argv[0] != "find" {
		t.Fatalf("entry=%+v ok=%v", entry, ok)
	}
}

func TestCommandCacheSkipsSecretArgv(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "commands.json")
	s := testCommandStore(t, path, config.Default().Cache)
	ctx := context.Background()
	key := CommandKeyFor("use secret token", sampleProfileForCommand())

	if err := s.Put(ctx, key, &providers.CommandResponse{
		Argv:       []string{"curl", "-H", "Authorization: Bearer sk-abcdefghijklmnopqrst"},
		Shell:      "bash",
		Confidence: 0.9,
	}); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Lookup(ctx, key); ok {
		t.Fatal("secret-shaped argv must not be cached")
	}
}

func TestCommandCacheCorruptFileDegrades(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "commands.json")
	if err := os.WriteFile(path, []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := testCommandStore(t, path, config.Default().Cache)
	key := CommandKeyFor("x", sampleProfileForCommand())
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected empty store after corrupt file")
	}
}
