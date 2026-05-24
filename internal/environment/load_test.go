package environment

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrDetectCreatesProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)

	p, err := LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.OS == "" {
		t.Fatal("expected OS in profile")
	}

	path := filepath.Join(dir, "system_profile.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("profile not written: %v", err)
	}

	p2, err := LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p2.OS != p.OS {
		t.Fatalf("reload OS %q want %q", p2.OS, p.OS)
	}
}
