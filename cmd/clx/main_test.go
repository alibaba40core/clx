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

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderrWriter{t: t})
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("stdout=%q", stdout.String())
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

func TestRunDryRunByDefault(t *testing.T) {
	setupCLIHome(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"pwd"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("expected dry-run line, stdout=%q", stdout.String())
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
