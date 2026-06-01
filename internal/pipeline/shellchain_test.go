package pipeline

import (
	"bytes"
	"context"
	"runtime"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestRunChainedShellDryRun(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows cmd chain test")
	}
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	testProfile(t, "windows", "cmd")

	var stdout, stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), `echo hello | findstr hello`, Options{
		DryRun: true,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("dry-run:")) {
		t.Fatalf("stdout=%s", stdout.String())
	}
}
