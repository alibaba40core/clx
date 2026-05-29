package main

import (
	"bytes"
	"testing"
	"time"
)

func TestVersionExitFast(t *testing.T) {
	if testing.Short() {
		t.Skip("timing smoke skipped in -short mode")
	}
	t.Setenv("CLX_HOME", t.TempDir())
	start := time.Now()
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)
	elapsed := time.Since(start)
	if code != 0 {
		t.Fatalf("exit %d stderr=%s", code, stderr.String())
	}
	// Generous local ceiling; CI enforces stricter limits via scripts/check-budgets.sh.
	if elapsed > 500*time.Millisecond {
		t.Fatalf("--version took %v", elapsed)
	}
}
