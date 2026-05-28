package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/policy"
)

// Run executes gen via direct argv or a validated shell host script.
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

	runCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd, err := buildCommand(runCtx, gen, cfg.Profile)
	if err != nil {
		return err
	}
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return &TimeoutError{After: cfg.Timeout}
		}
		return err
	}
	return nil
}
