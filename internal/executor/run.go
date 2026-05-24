package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/policy"
)

// Run executes gen.Argv via exec.CommandContext (argv-only, no shell).
func Run(ctx context.Context, gen generator.GeneratedCommand, opts ...Option) error {
	cfg := RunConfig{
		Timeout: 30 * time.Second,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
	for _, o := range opts {
		o(&cfg)
	}

	if !cfg.HasRisk {
		return ErrMissingRisk
	}
	if !cfg.HasPolicy {
		return ErrMissingPolicy
	}
	if !cfg.Policy.Allowed {
		return fmt.Errorf("%w: %s", policy.ErrBlocked, cfg.Policy.Reason)
	}
	if len(gen.Argv) == 0 {
		return ErrEmptyArgv
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	bin := gen.Argv[0]
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("command not found: %q", bin)
	}

	runCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, gen.Argv[0], gen.Argv[1:]...)
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
