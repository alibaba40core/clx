package test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/pipeline"
	"github.com/alibaba40core/clx/internal/policy"
)

func setupCLXHome(t *testing.T, osName, shell string, tools []string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(environment.SystemProfile{
		OS:             osName,
		Shell:          shell,
		AvailableTools: tools,
	})
	if err := environment.SaveStore(context.Background(), path, store); err != nil {
		t.Fatal(err)
	}
}

func TestE2EExplainSeedIntentsLinux(t *testing.T) {
	setupCLXHome(t, "linux", "bash", []string{"grep"})

	cases := []struct {
		input  string
		wantIn string
	}{
		{"grep errors logs.txt", "search_text_in_file"},
		{"locate help.txt", "find_file"},
		{"ls .", "list_dir"},
		{"pwd", "current_dir"},
		{"disk usage", "disk_usage"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code, err := pipeline.Run(context.Background(), config.Default(), tc.input, pipeline.Options{
				Explain: true,
				Stdout:  &stdout,
				Stderr:  &stderr,
			})
			if err != nil || code != 0 {
				t.Fatalf("code=%d err=%v stderr=%s", code, err, stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.wantIn) {
				t.Fatalf("stdout=%q", stdout.String())
			}
		})
	}
}

func TestE2EExplainSeedIntentsWindows(t *testing.T) {
	setupCLXHome(t, "windows", "powershell", nil)

	var stdout bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "grep errors logs.txt", pipeline.Options{
		Explain: true,
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Select-String") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestE2EExecPwdWithYes(t *testing.T) {
	if _, err := exec.LookPath("pwd"); err != nil {
		t.Skip("pwd not in PATH")
	}
	setupCLXHome(t, "linux", "bash", nil)

	var stdout, stderr bytes.Buffer
	input := "pwd"
	cfg := config.Default()
	cfg.Safety.DryRun = false // proves real exec path; opt out of new default
	code, err := pipeline.Run(context.Background(), cfg, input, pipeline.Options{
		Yes:    true,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("stderr=%s err=%v", stderr.String(), err)
	}
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}

func TestE2EDryRunFromDefaultConfig(t *testing.T) {
	setupCLXHome(t, "linux", "bash", nil)

	var stdout, stderr bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "pwd", pipeline.Options{
		Yes:    true,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "dry-run:") {
		t.Fatalf("expected dry-run on default config, stdout=%q", stdout.String())
	}
}

func TestE2EProfileWritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	p, err := environment.LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.OS == "" {
		t.Fatal("expected detect to fill OS")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var store environment.ProfileStore
	if err := json.Unmarshal(data, &store); err != nil {
		t.Fatal(err)
	}
	for _, prof := range store.Profiles {
		if prof.OS != "" {
			return
		}
	}
	t.Fatal("profile store has no profiles with OS")
}
