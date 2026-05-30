package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
)

func TestRunVersion(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "clx version") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunDoctor(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run([]string{"doctor"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "environment profile written") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunDoctorRefresh(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run([]string{"doctor", "--refresh"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "environment profile written") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunDoctorRefreshShortFlag(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run([]string{"doctor", "-r"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d, stderr=%s", code, stderr.String())
	}
}

func TestRunProviderFlagInvalidValue(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	var stderr bytes.Buffer
	code := run([]string{"--provider", "bogus", "pwd"}, &bytes.Buffer{}, &stderr)
	if code != 2 {
		t.Fatalf("exit %d, want 2, stderr=%s", code, stderr.String())
	}
}

func TestRunProviderOpenAIMissingKey(t *testing.T) {
	setupCLIHome(t)
	var stderr bytes.Buffer
	code := run([]string{"--provider", "openai", "find logs"}, &bytes.Buffer{}, &stderr)
	if code != 2 {
		t.Fatalf("exit %d, want 2, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "api_key") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderrWriter{t: t})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "clx alias") {
		t.Fatalf("help should document alias subcommand, stdout=%q", stdout.String())
	}
}

func TestRunPolicyListNotPipeline(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"policy", "list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "ai-generated command") {
		t.Fatalf("policy list must not run AI pipeline, stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Allowed verbs") {
		t.Fatalf("list should label output, stdout=%q", stdout.String())
	}
}

func TestRunPolicyShowNotPipeline(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"policy", "show"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Blocked patterns") {
		t.Fatalf("show should list blocks, stdout=%q", stdout.String())
	}
	if strings.Contains(stdout.String(), "ai-generated command") {
		t.Fatalf("policy show must not run AI pipeline, stdout=%q", stdout.String())
	}
}

func TestRunAliasListNotPipeline(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"alias", "list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "ai-generated command") {
		t.Fatalf("alias list must not run AI pipeline, stdout=%q", stdout.String())
	}
}

func setupCLIHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	p, err := environment.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(p)
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

func TestRunMediumModeDefaultNoDryRun(t *testing.T) {
	setupCLIHome(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"--explain", "pwd"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("default medium should not dry-run low risk, stdout=%q", stdout.String())
	}
}

func TestRunExplainGrep(t *testing.T) {
	setupCLIHome(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"--explain", "grep", "errors", "logs.txt"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "search_text_in_file") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

type stderrWriter struct{ t *testing.T }

func (w stderrWriter) Write(p []byte) (int, error) {
	w.t.Logf("stderr: %s", p)
	return len(p), nil
}
