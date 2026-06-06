//go:build lite

package pipeline

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestRunLiteExplainPwd(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "", "")

	var stdout bytes.Buffer
	cfg := config.Default()
	cfg.Provider = "none"
	cfg.Providers.Primary = "none"
	code, err := Run(context.Background(), cfg, "pwd", Options{
		Explain: true,
		Stdout:  &stdout,
		Stderr:  &bytes.Buffer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(stdout.String(), "Source:      Rule") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunLiteNotFoundNLNoAI(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "", "")

	var stderr bytes.Buffer
	cfg := config.Default()
	cfg.Features.AICommandGeneration = false
	cfg.Provider = "none"
	cfg.Providers.Primary = "none"
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

func TestRunLiteDelegatesWhenAIMissing(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "", "")

	var stderr bytes.Buffer
	cfg := config.Default()
	cfg.Provider = "ollama"
	cfg.Providers.Primary = "ollama"
	code, err := Run(context.Background(), cfg, "totally unknown nl phrase xyz", Options{
		ForwardedArgv: []string{"--provider", "ollama", "totally unknown nl phrase xyz"},
		Stderr:        &stderr,
		Stdout:        &bytes.Buffer{},
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "clx-ai not found") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
