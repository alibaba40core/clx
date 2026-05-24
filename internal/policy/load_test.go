package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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
