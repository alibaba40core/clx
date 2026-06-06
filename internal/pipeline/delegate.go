//go:build lite

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/alibaba40core/clx/internal/clxsidecar"
)

const workerEnvKey = "CLX_WORKER"

// delegateToWorker re-execs the hidden clx-ai sibling with the user's original argv.
func delegateToWorker(ctx context.Context, opts *Options) (int, error) {
	workerPath := opts.WorkerPath
	if workerPath == "" {
		var err error
		workerPath, err = clxsidecar.WorkerPath()
		if err != nil {
			fmt.Fprintf(opts.Stderr, "AI support is not installed (clx-ai not found next to clx): %v\nReinstall CLX or run make install.\n", err)
			return 1, err
		}
	}

	argv := opts.ForwardedArgv
	if len(argv) == 0 {
		return 1, errors.New("no arguments to forward to AI worker")
	}

	cmd := exec.CommandContext(ctx, workerPath, argv...)
	cmd.Env = append(os.Environ(), workerEnvKey+"=1")
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code := exitErr.ExitCode()
			if code < 0 {
				return 1, err
			}
			return code, err
		}
		fmt.Fprintf(opts.Stderr, "AI worker: %v\n", err)
		return 1, err
	}
	return 0, nil
}
