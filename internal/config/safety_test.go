package config

import (
	"strings"
	"testing"
)

func TestDecideSafetyActionPresetMatrix(t *testing.T) {
	cfg := Default()
	flags := SafetyFlagOverrides{}

	cases := []struct {
		mode, risk string
		want       SafetyAction
	}{
		{"low", "low", SafetyAction{}},
		{"low", "medium", SafetyAction{}},
		{"low", "high", SafetyAction{Confirm: true}},
		{"medium", "low", SafetyAction{}},
		{"medium", "medium", SafetyAction{ShowExplain: true, Confirm: true}},
		{"medium", "high", SafetyAction{ShowExplain: true, Confirm: true}},
		{"high", "low", SafetyAction{ShowExplain: true, Confirm: true}},
		{"high", "medium", SafetyAction{ShowExplain: true, Preview: true, Confirm: true, BlockYes: true}},
		{"high", "high", SafetyAction{ShowExplain: true, Preview: true, Confirm: true, BlockYes: true}},
	}
	for _, tc := range cases {
		cfg.Safety.Mode = tc.mode
		got := DecideSafetyAction(cfg, tc.risk, flags)
		if got != tc.want {
			t.Errorf("mode=%s risk=%s: got %+v want %+v", tc.mode, tc.risk, got, tc.want)
		}
	}
}

func TestDecideSafetyActionCustom(t *testing.T) {
	cfg := Default()
	cfg.Safety.Mode = "custom"
	cfg.Safety.RequireConfirmation = true
	cfg.Safety.DryRun = true
	cfg.Features.Explain = false

	got := DecideSafetyAction(cfg, "low", SafetyFlagOverrides{})
	if got.Preview != true || got.Confirm != true || got.ShowExplain != false {
		t.Fatalf("custom toggles: got %+v", got)
	}
}

func TestDecideSafetyActionDryRunFlag(t *testing.T) {
	cfg := Default()
	cfg.Safety.Mode = "low"
	got := DecideSafetyAction(cfg, "high", SafetyFlagOverrides{DryRun: true})
	if !got.Preview || got.Confirm {
		t.Fatalf("--dry-run should preview only: got %+v", got)
	}
}

func TestPreviewOnlyCustomDryRun(t *testing.T) {
	cfg := Default()
	cfg.Safety.Mode = "custom"
	cfg.Safety.DryRun = true
	cfg.Safety.RequireConfirmation = false
	action := DecideSafetyAction(cfg, "low", SafetyFlagOverrides{})
	if !action.PreviewOnly(cfg, SafetyFlagOverrides{}) {
		t.Fatal("custom dry_run without confirm should be preview-only")
	}
}

func TestShouldConfirmBlockYes(t *testing.T) {
	cfg := Default()
	action := SafetyAction{Confirm: true, BlockYes: true}
	if !action.ShouldConfirm(cfg, SafetyFlagOverrides{Yes: true}) {
		t.Fatal("BlockYes should require confirm even with -y")
	}
	action2 := SafetyAction{Confirm: true, BlockYes: false}
	if action2.ShouldConfirm(cfg, SafetyFlagOverrides{Yes: true}) {
		t.Fatal("unblocked -y should skip confirm")
	}
}

func TestShouldConfirmBlockYesOverAutoExecute(t *testing.T) {
	cfg := Default()
	cfg.Execution.AutoExecute = true
	action := SafetyAction{Confirm: true, BlockYes: true}
	if !action.ShouldConfirm(cfg, SafetyFlagOverrides{Yes: true}) {
		t.Fatal("BlockYes should require confirm even with auto_execute and -y")
	}
}

func TestShouldConfirmAutoExecuteSkipsWhenNotBlocked(t *testing.T) {
	cfg := Default()
	cfg.Execution.AutoExecute = true
	action := SafetyAction{Confirm: true, BlockYes: false}
	if action.ShouldConfirm(cfg, SafetyFlagOverrides{}) {
		t.Fatal("auto_execute should skip confirm when BlockYes is unset")
	}
}

func TestApplySafetyMode(t *testing.T) {
	cfg := Default()
	if err := ApplySafetyMode(&cfg, "high"); err != nil || cfg.Safety.Mode != "high" {
		t.Fatalf("ApplySafetyMode high: mode=%q err=%v", cfg.Safety.Mode, err)
	}
	if err := ApplySafetyMode(&cfg, "bad"); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestApplySafetyCustomToggle(t *testing.T) {
	cfg := Default()
	if err := ApplySafetyCustomToggle(&cfg, "dry_run", true); err != nil {
		t.Fatal(err)
	}
	if cfg.Safety.Mode != "custom" || !cfg.Safety.DryRun {
		t.Fatalf("cfg=%+v", cfg.Safety)
	}
}

func TestSafetyShowLinesPreset(t *testing.T) {
	cfg := Default()
	cfg.Safety.Mode = "low"
	lines := SafetyShowLines(cfg)
	text := strings.Join(lines, "\n")
	for _, want := range []string{
		"Safety mode: low",
		"Command risk",
		"high:  confirm",
		"clx safety set mode=",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in\n%s", want, text)
		}
	}
}

func TestSafetyShowLinesMedium(t *testing.T) {
	cfg := Default()
	lines := SafetyShowLines(cfg)
	text := strings.Join(lines, "\n")
	if !strings.Contains(text, "Balanced (default)") || !strings.Contains(text, "explain + confirm") {
		t.Fatalf("lines=%q", text)
	}
}

func TestSafetyShowLinesCustom(t *testing.T) {
	cfg := Default()
	cfg.Safety.Mode = "custom"
	cfg.Safety.DryRun = true
	lines := SafetyShowLines(cfg)
	text := strings.Join(lines, "\n")
	if !strings.Contains(text, "dry_run:") || !strings.Contains(text, "Custom toggles") {
		t.Fatalf("lines=%q", text)
	}
}
