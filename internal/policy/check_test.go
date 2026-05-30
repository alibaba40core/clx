package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

func TestCheckBlocked(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte("blocked:\n  - \"rm -rf\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "/"}, Command: "rm -rf /"}
	got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecAllowed {
		t.Fatal("expected blocked")
	}
}

func TestCheckExplainOnlyAllowListWarns(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte("allowed:\n  - git\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	gen := generator.GeneratedCommand{Argv: []string{"pwd"}, Command: "pwd"}
	got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{
		SafetyMode:  "high",
		ExplainOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Allowed || got.ExecAllowed {
		t.Fatalf("explain should warn only: %+v", got)
	}
	if got.Reason != "command verb not on allow list" {
		t.Fatalf("reason=%q", got.Reason)
	}
}

func TestCheckBlockArgvAware(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	policyYAML := `blocked:
  - "format"
  - "rm -rf"
  - "shutdown"
`
	if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte(policyYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Run("format_no_false_positive_in_path", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{
			Argv:    []string{"find", "./form", "-name", "x"},
			Command: "find ./form -name x",
		}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !got.ExecAllowed {
			t.Fatalf("unexpected block: %s", got.Reason)
		}
	})

	t.Run("format_blocks_exact_token", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"format", "C:"}, Command: "format C:"}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ExecAllowed {
			t.Fatal("expected blocked for format verb")
		}
	})

	t.Run("rm_rf_subsequence", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "."}, Command: "rm -rf ."}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ExecAllowed {
			t.Fatal("expected blocked")
		}
	})

	t.Run("rm_rf_concatenated_not_blocked", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"rm-rf", "."}, Command: "rm-rf ."}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !got.ExecAllowed {
			t.Fatalf("rm-rf single token must not match rm -rf pattern: %s", got.Reason)
		}
	})

	t.Run("shutdown_blocks", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"shutdown", "/s"}, Command: "shutdown /s"}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ExecAllowed {
			t.Fatal("expected blocked")
		}
	})
}

func TestCheckAllowList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "allowed:\n  - git\n  - docker\n"
	if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	high := CheckOptions{SafetyMode: "high"}

	t.Run("allowed_verb_passes_high", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"git", "status"}, Command: "git status"}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, high)
		if err != nil {
			t.Fatal(err)
		}
		if !got.ExecAllowed {
			t.Fatalf("expected allowed: %s", got.Reason)
		}
	})

	t.Run("disallowed_verb_blocked_high", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"npm", "install"}, Command: "npm install"}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, high)
		if err != nil {
			t.Fatal(err)
		}
		if got.ExecAllowed {
			t.Fatal("expected blocked by allow list")
		}
	})

	t.Run("medium_ignores_allow_list", func(t *testing.T) {
		t.Parallel()
		gen := generator.GeneratedCommand{Argv: []string{"pwd"}, Command: "pwd"}
		got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{SafetyMode: "medium"})
		if err != nil {
			t.Fatal(err)
		}
		if !got.ExecAllowed {
			t.Fatalf("medium must ignore allow list: %s", got.Reason)
		}
	})
}

func TestCheckAccessLevel(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}

	writePolicy := func(yaml string) {
		t.Helper()
		ResetCache()
		if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte(yaml), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	genLow := generator.GeneratedCommand{Argv: []string{"grep", "x"}, Command: "grep x"}
	genMed := generator.GeneratedCommand{Argv: []string{"mkdir", "x"}, Command: "mkdir x"}
	raLow := risk.RiskAssessment{Level: risk.Low}
	raMed := risk.RiskAssessment{Level: risk.Medium}

	t.Run("safe_denies_all", func(t *testing.T) {
		writePolicy("access_level: safe\n")
		got, err := Check(context.Background(), genLow, raLow, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ExecAllowed {
			t.Fatal("safe must deny execution")
		}
	})

	t.Run("moderate_allows_low", func(t *testing.T) {
		writePolicy("access_level: moderate\n")
		got, err := Check(context.Background(), genLow, raLow, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !got.ExecAllowed {
			t.Fatalf("moderate should allow low: %s", got.Reason)
		}
	})

	t.Run("moderate_denies_medium", func(t *testing.T) {
		writePolicy("access_level: moderate\n")
		got, err := Check(context.Background(), genMed, raMed, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ExecAllowed {
			t.Fatal("moderate must deny medium risk")
		}
	})

	t.Run("full_allows_medium", func(t *testing.T) {
		writePolicy("access_level: full\n")
		got, err := Check(context.Background(), genMed, raMed, CheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !got.ExecAllowed {
			t.Fatalf("full should allow: %s", got.Reason)
		}
	})
}

func TestCheckAllowListEmptyMeansNoGate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte("blocked:\n  - shutdown\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	gen := generator.GeneratedCommand{Argv: []string{"npm", "install"}, Command: "npm install"}
	got, err := Check(context.Background(), gen, risk.RiskAssessment{}, CheckOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.ExecAllowed {
		t.Fatalf("empty allow list must not gate: %s", got.Reason)
	}
}

func TestArgvMatchesBlocked(t *testing.T) {
	t.Parallel()
	cases := []struct {
		argv    []string
		pattern string
		want    bool
	}{
		{[]string{"rm", "-rf", "/"}, "rm -rf", true},
		{[]string{"find", "./form"}, "format", false},
		{[]string{"format", "C:"}, "format", true},
		{[]string{"rm-rf"}, "rm -rf", false},
	}
	for _, tc := range cases {
		got := argvMatchesBlocked(tc.argv, tokenizePattern(tc.pattern))
		if got != tc.want {
			t.Fatalf("argv %v pattern %q: got %v want %v", tc.argv, tc.pattern, got, tc.want)
		}
	}
}
