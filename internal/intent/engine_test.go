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

func TestKnownIntentsIncludesBuiltIn(t *testing.T) {
	eng, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	names := eng.KnownIntents()
	if len(names) == 0 {
		t.Fatal("expected built-in intents")
	}
	found := false
	for _, n := range names {
		if n == "find_file" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("find_file not in %v", names)
	}
}

func TestKnownIntentsOverlay(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	rulesDir := filepath.Join(dir, "rules")
	if err := os.MkdirAll(rulesDir, 0o700); err != nil {
		t.Fatal(err)
	}
	overlay := `intent: custom_intent
examples:
  - do custom {{thing}}
strategies:
  default:
    argv: [echo, ok]
`
	if err := os.WriteFile(filepath.Join(rulesDir, "custom.yaml"), []byte(overlay), 0o600); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngineWithOverlay(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range eng.KnownIntents() {
		if n == "custom_intent" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("overlay intent missing from KnownIntents")
	}
}

func TestValidateResolvedAcceptsKnownIntent(t *testing.T) {
	eng, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	err = eng.ValidateResolved(ResolvedIntent{
		Intent: "find_file",
		Params: map[string]string{"filename": "help.txt"},
		Source: SourceAI,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateResolvedRejectsUnknownIntent(t *testing.T) {
	eng, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	err = eng.ValidateResolved(ResolvedIntent{
		Intent: "rm_rf_slash",
		Params: map[string]string{},
		Source: SourceAI,
	})
	if err == nil {
		t.Fatal("expected error for unknown intent")
	}
}

func TestValidateResolvedRejectsExtraParam(t *testing.T) {
	eng, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	err = eng.ValidateResolved(ResolvedIntent{
		Intent: "find_file",
		Params: map[string]string{
			"filename":   "x",
			"evil_extra": "y",
		},
		Source: SourceAI,
	})
	if err == nil {
		t.Fatal("expected error for extra param")
	}
	if !strings.Contains(err.Error(), "evil_extra") {
		t.Fatalf("err=%v", err)
	}
}

func TestValidateResolvedRejectsMissingDeclaredParam(t *testing.T) {
	eng := NewEngine([]Rule{{
		Intent:   "needs_both",
		Examples: []string{"go {{a}} {{b}}"},
		Params:   []string{"a", "b"},
		Strategies: map[string]Strategy{
			"default": {Argv: []string{"echo", "{{a}}", "{{b}}"}},
		},
	}})
	err := eng.ValidateResolved(ResolvedIntent{
		Intent: "needs_both",
		Params: map[string]string{"a": "1"},
		Source: SourceAI,
	})
	if err == nil {
		t.Fatal("expected missing param error")
	}
	if !strings.Contains(err.Error(), "missing param") {
		t.Fatalf("err=%v", err)
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
