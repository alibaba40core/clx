package main

import (
	"bytes"
	"strings"
	"testing"
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

func TestRunExplainGrep(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
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
