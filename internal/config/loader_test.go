package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDefault(t *testing.T) {
	t.Parallel()
	if err := Validate(Default()); err != nil {
		t.Fatal(err)
	}
}

func TestValidateInvalidProvider(t *testing.T) {
	t.Parallel()
	c := Default()
	c.Provider = "invalid"
	if err := Validate(c); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadMergesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("provider: openai\nmodel: gpt-test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != "openai" || cfg.Model != "gpt-test" {
		t.Fatalf("got %+v", cfg)
	}
	if cfg.Execution.Timeout != 30 {
		t.Fatal("expected default timeout merged")
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	t.Parallel()
	cfg, err := Load(context.Background(), filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != "ollama" {
		t.Fatalf("got %+v", cfg)
	}
}
