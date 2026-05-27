package executor

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func buildCommand(ctx context.Context, gen generator.GeneratedCommand, profile environment.SystemProfile) (*exec.Cmd, error) {
	host := gen.ExecHost
	if host == generator.ExecDirect {
		if len(gen.Argv) == 0 {
			return nil, ErrEmptyArgv
		}
		if _, err := exec.LookPath(gen.Argv[0]); err != nil {
			if runtime.GOOS == "windows" {
				return nil, fmt.Errorf("command not found: %q (try clx --explain)", gen.Argv[0])
			}
			host = generator.ExecPosix
		}
	}
	return buildCommandForHost(ctx, host, gen, profile)
}

func buildCommandForHost(ctx context.Context, host generator.ExecHost, gen generator.GeneratedCommand, profile environment.SystemProfile) (*exec.Cmd, error) {
	if len(gen.Argv) == 0 {
		return nil, ErrEmptyArgv
	}

	switch host {
	case generator.ExecDirect:
		bin := gen.Argv[0]
		if _, err := exec.LookPath(bin); err != nil {
			return nil, fmt.Errorf("command not found: %q", bin)
		}
		return exec.CommandContext(ctx, gen.Argv[0], gen.Argv[1:]...), nil

	case generator.ExecPowerShell:
		exe, err := ResolvePowerShell(profile)
		if err != nil {
			return nil, err
		}
		script, err := BuildValidatedScript("powershell", gen.Argv)
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, exe, "-NoProfile", "-NonInteractive", "-Command", script), nil

	case generator.ExecCmd:
		exe, err := ResolveCmd()
		if err != nil {
			return nil, err
		}
		script, err := BuildValidatedScript("cmd", gen.Argv)
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, exe, "/c", script), nil

	case generator.ExecPosix:
		exe, err := ResolvePosixShell()
		if err != nil {
			return nil, err
		}
		shellName := shellNameForScript(host, profile)
		script, err := BuildValidatedScript(shellName, gen.Argv)
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, exe, "-c", script), nil

	default:
		return nil, fmt.Errorf("executor: unknown exec host %d", host)
	}
}
