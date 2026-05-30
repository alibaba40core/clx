package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestAllowListActive(t *testing.T) {
	t.Parallel()
	allowed := []string{"git"}
	if !AllowListActive("high", allowed) {
		t.Fatal("high + non-empty allowed should be active")
	}
	if AllowListActive("high", nil) {
		t.Fatal("high + empty allowed should not gate")
	}
	if AllowListActive("medium", allowed) {
		t.Fatal("medium should ignore allow list")
	}
}

func TestLoadSnapshot(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	p, err := config.PolicyPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("access_level: full\nblocked:\n  - shutdown\nallowed:\n  - git\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ResetCache()

	snap, err := LoadSnapshot(context.Background(), "medium")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.FileExists || len(snap.Blocked) != 1 || len(snap.Allowed) != 1 {
		t.Fatalf("snap=%+v", snap)
	}
	if AllowListActive(snap.SafetyMode, snap.Allowed) {
		t.Fatal("medium must not activate allow list")
	}
}
