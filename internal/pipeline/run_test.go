package pipeline

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/policy"
)

func testProfile(t *testing.T, osName, shell string) {
	t.Helper()
	if osName == "" || shell == "" {
		mp := environment.MinimalProfile()
		osName = mp.OS
		shell = mp.Shell
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(environment.SystemProfile{
		OS:             osName,
		Shell:          shell,
		AvailableTools: []string{"grep"},
	})
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

func TestRunExplainGrepLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux profile test")
	}
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout, stderr bytes.Buffer
	cfg := config.Default()
	code, err := Run(context.Background(), cfg, "grep errors logs.txt", Options{
		Explain: true,
		Stdout:  &stdout,
		Stderr:  &stderr,
	})
	if err != nil {
		t.Fatalf("stderr=%s err=%v", stderr.String(), err)
	}
	if code != 0 {
		t.Fatalf("code %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "search_text_in_file") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "grep") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunDryRun(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "pwd", Options{
		DryRun: true,
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stdout=%s", code, err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunExplainHighAllowListWarns(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	polPath, err := config.PolicyPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(polPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(polPath, []byte("allowed:\n  - git\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()

	cfg := config.Default()
	cfg.Safety.Mode = "high"
	var stdout bytes.Buffer
	code, err := Run(context.Background(), cfg, "pwd", Options{
		Explain: true,
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stdout.String(), "Policy (exec):") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunMediumModeLowRiskNoDryRun(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "pwd", Options{
		Explain: true,
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stdout=%s", code, err, stdout.String())
	}
	if strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("medium mode low risk should not dry-run, stdout=%q", stdout.String())
	}
}

func TestRunCustomDryRunPreviewOnly(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	cfg := config.Default()
	cfg.Safety.Mode = "custom"
	cfg.Safety.DryRun = true
	cfg.Safety.RequireConfirmation = false

	var stdout bytes.Buffer
	code, err := Run(context.Background(), cfg, "pwd", Options{
		Yes:    true,
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stdout=%s", code, err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("custom dry_run should preview, stdout=%q", stdout.String())
	}
}

func TestRunDryRunFromConfig(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout bytes.Buffer
	cfg := config.Default()
	cfg.Safety.Mode = "custom"
	cfg.Safety.DryRun = true
	code, err := Run(context.Background(), cfg, "pwd", Options{
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stdout=%s", code, err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunYesDoesNotBypassConfigDryRun(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout bytes.Buffer
	cfg := config.Default()
	cfg.Safety.Mode = "custom"
	cfg.Safety.DryRun = true
	code, err := Run(context.Background(), cfg, "pwd", Options{
		Yes:    true,
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stdout=%s", code, err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("expected dry-run with -y on default config, stdout=%q", stdout.String())
	}
}

func TestRunDryRunFlagWhenConfigDisabled(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	cfg := config.Default()
	cfg.Safety.DryRun = false

	var stdout bytes.Buffer
	code, err := Run(context.Background(), cfg, "pwd", Options{
		DryRun: true,
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stdout=%s", code, err, stdout.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunNilAIResolverSameAsRulesOnly(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout, stderr bytes.Buffer
	cfg := config.Default()
	code, err := Run(context.Background(), cfg, "pwd", Options{
		Explain:    true,
		AIResolver: nil,
		Stdout:     &stdout,
		Stderr:     &stderr,
	})
	if err != nil {
		t.Fatalf("stderr=%s err=%v", stderr.String(), err)
	}
	if code != 0 {
		t.Fatalf("code %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "current_dir") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Source:      Rule") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

type fakeAIResolver struct {
	result intent.ResolvedIntent
	err    error
}

func (f *fakeAIResolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	if f.err != nil {
		return intent.ResolvedIntent{}, f.err
	}
	return f.result, nil
}

func TestRunAIResolverRejectsExtraParam(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "totally unknown input xyz", Options{
		AIResolver: &fakeAIResolver{
			result: intent.ResolvedIntent{
				Intent: "find_file",
				Params: map[string]string{
					"filename":   "x",
					"evil_extra": "y",
				},
				Source: intent.SourceAI,
			},
		},
		Stderr: &stderr,
		Stdout: &bytes.Buffer{},
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "untrusted resolver output rejected") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunAIResolverRejectsUnknownIntent(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "unknown xyz", Options{
		AIResolver: &fakeAIResolver{
			result: intent.ResolvedIntent{
				Intent: "rm_rf_slash",
				Params: map[string]string{},
				Source: intent.SourceAI,
			},
		},
		Stderr: &stderr,
		Stdout: &bytes.Buffer{},
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "untrusted resolver output rejected") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunAIResolverValidIntentExplain(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "", "")

	var stdout, stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "unknown xyz", Options{
		Explain: true,
		AIResolver: &fakeAIResolver{
			result: intent.ResolvedIntent{
				Intent: "current_dir",
				Params: map[string]string{},
				Source: intent.SourceAI,
			},
		},
		Logger: slog.Default(),
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("stderr=%s err=%v", stderr.String(), err)
	}
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Source:      AI") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "current_dir") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunAIMissNaturalLanguage(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "", "")

	var stderr bytes.Buffer
	cfg := config.Default()
	cfg.Features.AICommandGeneration = false
	cfg.Provider = "none"
	code, err := Run(context.Background(), cfg, "find all widgets modified yesterday", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Stderr:     &stderr,
		Stdout:     &bytes.Buffer{},
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "AI could not map") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunNotFoundNL(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "", "")

	var stderr bytes.Buffer
	cfg := config.Default()
	cfg.Features.AICommandGeneration = false
	cfg.Provider = "none"
	code, err := Run(context.Background(), cfg, "find all widgets modified yesterday", Options{
		Stderr: &stderr,
		Stdout: &bytes.Buffer{},
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "natural language") && !strings.Contains(stderr.String(), "no matching") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
