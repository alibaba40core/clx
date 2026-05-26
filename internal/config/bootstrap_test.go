package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapCreatesTree(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)

	result, err := Bootstrap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.WroteConfig || !result.WrotePolicy {
		t.Fatalf("expected writes on first run: %+v", result)
	}

	mustExist(t, filepath.Join(home, "config.yaml"))
	mustExist(t, filepath.Join(home, "policies", "policy.yaml"))
	mustExist(t, filepath.Join(home, "cache"))
	mustExist(t, filepath.Join(home, "logs"))
}

func TestBootstrapIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)

	if _, err := Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(home, "config.yaml")
	custom := []byte("provider: openai\nmodel: custom\n")
	if err := os.WriteFile(cfgPath, custom, 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Bootstrap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.WroteConfig {
		t.Fatal("must not overwrite existing config")
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(custom) {
		t.Fatalf("config was modified: %s", data)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}
