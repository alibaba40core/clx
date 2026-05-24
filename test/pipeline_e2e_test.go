package test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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
	p := environment.SystemProfile{
		SchemaVersion:  environment.SchemaVersion,
		OS:             osName,
		Shell:          shell,
		AvailableTools: tools,
	}
	if err := environment.Save(context.Background(), path, p); err != nil {
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
	code, err := pipeline.Run(context.Background(), config.Default(), input, pipeline.Options{
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

func TestE2EProfileWritten(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	path := filepath.Join(dir, "system_profile.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
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
	var decoded environment.SystemProfile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.OS == "" {
		t.Fatal("profile file still empty")
	}
}
