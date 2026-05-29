package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestRunConfigShow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	code := run([]string{"config", "show"}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout.String(), "provider: ollama") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunConfigSetStdinEncrypts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	secret := "sk-teststdinkey1234567890"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })
	go func() {
		_, _ = w.WriteString(secret)
		_ = w.Close()
	}()

	var stdout, stderr bytes.Buffer
	code := run([]string{"config", "set", "providers.openai.api_key", "--stdin"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}

	cfgPath, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatal("plaintext secret written to config")
	}
	if !strings.Contains(string(raw), "enc:v1:") {
		t.Fatalf("config=%s", raw)
	}
}

func TestRunConfigShowNeverPrintsPlaintextKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	path := filepath.Join(home, "config.yaml")
	cfg := config.Default()
	cfg.Providers.OpenAI.APIKey = "sk-leak-test-key-abcdefghij"
	if err := config.Save(context.Background(), path, cfg); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	code := run([]string{"config", "show", "--config", path}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	out := stdout.String()
	if strings.Contains(out, "sk-leak-test-key") {
		t.Fatalf("leaked secret in show output: %s", out)
	}
	if !strings.Contains(out, "****") {
		t.Fatalf("expected masked secret: %s", out)
	}
}

func TestRunConfigProviderUseOpenAIWithKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	secret := "sk-providerusekey1234567890"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })
	go func() {
		_, _ = w.WriteString(secret)
		_ = w.Close()
	}()
	if code := run([]string{"config", "set", "providers.openai.api_key", "--stdin"}, &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatalf("set key exit %d", code)
	}

	var stdout bytes.Buffer
	code := run([]string{"config", "provider", "use", "openai"}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "active provider: openai") {
		t.Fatalf("stdout=%q", stdout.String())
	}

	cfgPath, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(context.Background(), cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != "openai" {
		t.Fatalf("provider=%q", cfg.Provider)
	}
}

func TestRunConfigProviderUseOpenAIRequiresKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	code := run([]string{"config", "provider", "use", "openai"}, &bytes.Buffer{}, &stderr)
	if code != 1 {
		t.Fatalf("exit %d, want 1, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "api_key") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunConfigProviderList(t *testing.T) {
	var stdout bytes.Buffer
	code := run([]string{"config", "provider", "list"}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout.String(), "ollama") || !strings.Contains(stdout.String(), "openai") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunConfigSetArgvRejectsSecret(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	code := run([]string{"config", "set", "providers.openai.api_key", "sk-argkey1234567890"}, &bytes.Buffer{}, &stderr)
	if code != 2 {
		t.Fatalf("exit %d, want 2, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "command line") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
