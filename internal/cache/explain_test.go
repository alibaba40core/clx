package cache

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/generator"
)

func TestExplainKeyForDeterministic(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Shell: "bash", Argv: []string{"ls", "-la"}}
	k1 := ExplainKeyFor("list_dir", gen)
	k2 := ExplainKeyFor("list_dir", gen)
	if k1 != k2 || k1 == "" {
		t.Fatalf("keys = %q %q", k1, k2)
	}
}

func TestExplainStoreMissHit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "explanations.json")
	cfg := config.Default().Cache
	s, err := LoadExplain(context.Background(), path, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	key := ExplainKeyFor("list_dir", generator.GeneratedCommand{Shell: "bash", Argv: []string{"ls"}})
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected miss")
	}
	if err := s.Put(context.Background(), key, "Lists directory contents."); err != nil {
		t.Fatal(err)
	}
	entry, ok := s.Lookup(context.Background(), key)
	if !ok || entry.Text != "Lists directory contents." {
		t.Fatalf("entry = %+v ok=%v", entry, ok)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "Lists directory contents.") {
		t.Fatalf("file=%q", raw)
	}
}

func TestExplainStoreCorruptStartsEmpty(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "explanations.json")
	if err := os.WriteFile(path, []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadExplain(context.Background(), path, config.Default().Cache, nil)
	if err != nil {
		t.Fatal(err)
	}
	key := ExplainKeyFor("x", generator.GeneratedCommand{Shell: "bash", Argv: []string{"y"}})
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected empty store")
	}
}
