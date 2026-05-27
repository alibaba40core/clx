package intent

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestNewDefaultEngineMatchesModuleRoot(t *testing.T) {
	embedded, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	fromFS, err := NewEngineFromModuleRoot()
	if err != nil {
		t.Skipf("module root not available: %v", err)
	}

	gotEmbedded := intentNames(embedded)
	gotFS := intentNames(fromFS)
	sort.Strings(gotEmbedded)
	sort.Strings(gotFS)

	if len(gotEmbedded) != len(gotFS) {
		t.Fatalf("intent count: embedded %d, module root %d", len(gotEmbedded), len(gotFS))
	}
	for i := range gotEmbedded {
		if gotEmbedded[i] != gotFS[i] {
			t.Fatalf("intent mismatch at %d: embedded %q, module root %q", i, gotEmbedded[i], gotFS[i])
		}
	}
}

func intentNames(eng *Engine) []string {
	names := make([]string, 0, len(eng.rules))
	for _, r := range eng.rules {
		names = append(names, r.Intent)
	}
	return names
}

func TestOverlayRuleOverridesBuiltIn(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	rulesDir := filepath.Join(dir, "rules")
	if err := os.MkdirAll(rulesDir, 0o700); err != nil {
		t.Fatal(err)
	}
	overlay := `intent: git_status
examples:
  - git status
strategies:
  default:
    argv:
      - echo
      - clx-overlay-marker
`
	if err := os.WriteFile(filepath.Join(rulesDir, "override.yaml"), []byte(overlay), 0o600); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngineWithOverlay(context.Background(), slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	rule, ok := eng.RuleForIntent("git_status")
	if !ok {
		t.Fatal("git_status missing")
	}
	strat := rule.Strategies["default"]
	if len(strat.Argv) < 2 || strat.Argv[0] != "echo" || strat.Argv[1] != "clx-overlay-marker" {
		t.Fatalf("overlay did not win: %+v", strat)
	}
}

func TestOverlayMissingDirsAreSilent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	embedded, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	withOverlay, err := NewEngineWithOverlay(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(intentNames(embedded)) != len(intentNames(withOverlay)) {
		t.Fatalf("counts differ: embedded %d overlay %d", len(embedded.rules), len(withOverlay.rules))
	}
}

func TestOverlayMalformedFileWarnsButDoesNotFail(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	rulesDir := filepath.Join(dir, "rules")
	if err := os.MkdirAll(rulesDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "bad.yaml"), []byte("{{not valid yaml"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	eng, err := NewEngineWithOverlay(context.Background(), logger)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := eng.RuleForIntent("git_status"); !ok {
		t.Fatal("built-in git_status missing after bad overlay file")
	}
	if !strings.Contains(buf.String(), "overlay rule file skipped") {
		t.Fatalf("expected warning in log, got %q", buf.String())
	}
}
