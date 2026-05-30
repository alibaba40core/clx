package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestRunSafetyShowDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	code := run([]string{"safety", "show"}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout.String(), "Safety mode: medium") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Command risk") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunSafetySetMode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	code := run([]string{"safety", "set", "mode=low"}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d stdout=%s", code, stdout.String())
	}

	cfgPath, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(context.Background(), cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Safety.Mode != "low" {
		t.Fatalf("mode=%q", cfg.Safety.Mode)
	}
}

func TestRunSafetySetToggleSwitchesCustom(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}

	code := run([]string{"safety", "set", "dry_run=true"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}

	cfgPath, err := config.ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "mode: custom") {
		t.Fatalf("config=%s", raw)
	}
}
