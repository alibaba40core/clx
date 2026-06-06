package clxsidecar_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alibaba40core/clx/internal/clxsidecar"
)

func TestWorkerPathAdjacentTo_Sibling(t *testing.T) {
	dir := t.TempDir()
	name := clxsidecar.WorkerBaseName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	worker := filepath.Join(dir, name)
	if err := os.WriteFile(worker, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	front := filepath.Join(dir, "clx")
	if runtime.GOOS == "windows" {
		front += ".exe"
	}

	got, err := clxsidecar.WorkerPathAdjacentTo(front)
	if err != nil {
		t.Fatal(err)
	}
	if got != worker {
		t.Fatalf("WorkerPathAdjacentTo() = %q, want %q", got, worker)
	}
}

func TestWorkerPathAdjacentTo_Missing(t *testing.T) {
	dir := t.TempDir()
	front := filepath.Join(dir, "clx")
	if runtime.GOOS == "windows" {
		front += ".exe"
	}
	if _, err := clxsidecar.WorkerPathAdjacentTo(front); err == nil {
		t.Fatal("expected error when worker missing")
	}
}
