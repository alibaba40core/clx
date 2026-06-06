package clxsidecar

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const WorkerBaseName = "clx-ai"

// WorkerPath returns the absolute path to the clx-ai worker binary installed
// next to the current clx executable.
func WorkerPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	return WorkerPathAdjacentTo(exe)
}

// WorkerPathAdjacentTo resolves the worker path given the front binary path.
func WorkerPathAdjacentTo(executable string) (string, error) {
	name := WorkerBaseName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(filepath.Dir(executable), name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return path, nil
}
