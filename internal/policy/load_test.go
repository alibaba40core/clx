package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadBlockedFromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	if err := os.MkdirAll(filepath.Join(dir, "policies"), 0o700); err != nil {
		t.Fatal(err)
	}
	content := "blocked:\n  - \"rm -rf\"\n"
	if err := os.WriteFile(filepath.Join(dir, "policies", "policy.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	f, err := Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Blocked) != 1 || f.Blocked[0] != "rm -rf" {
		t.Fatalf("blocked=%v", f.Blocked)
	}
}

func TestLoadReloadOnMtimeChange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(polDir, "policy.yaml")
	if err := os.WriteFile(path, []byte("blocked:\n  - alpha\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	f1, err := Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(f1.Blocked) != 1 || f1.Blocked[0] != "alpha" {
		t.Fatalf("first load blocked=%v", f1.Blocked)
	}

	// Ensure mtime advances on all platforms.
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(path, []byte("blocked:\n  - beta\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	f2, err := Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(f2.Blocked) != 1 || f2.Blocked[0] != "beta" {
		t.Fatalf("reload blocked=%v", f2.Blocked)
	}
}
