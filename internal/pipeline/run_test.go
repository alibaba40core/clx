package pipeline

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/policy"
)

func testProfile(t *testing.T, osName, shell string) {
	t.Helper()
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
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "linux", "bash")

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
	testProfile(t, "linux", "bash")

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

func TestRunNotFoundNL(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "linux", "bash")

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "find all files modified today", Options{
		Stderr: &stderr,
		Stdout: &bytes.Buffer{},
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "natural language") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
