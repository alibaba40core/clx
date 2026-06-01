package executor

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

// effectiveExecHost resolves the subprocess host after PATH lookups. Generator
// sets ExecHost from strategy key only; promotion and fallbacks live here.
func effectiveExecHost(gen generator.GeneratedCommand, profile environment.SystemProfile) generator.ExecHost {
	host := gen.ExecHost
	if host == generator.ExecCmd && len(gen.Argv) > 0 {
		if _, err := exec.LookPath(gen.Argv[0]); err == nil {
			return generator.ExecDirect
		}
	}
	if host == generator.ExecDirect && len(gen.Argv) > 0 {
		if _, err := exec.LookPath(gen.Argv[0]); err != nil {
			if runtime.GOOS == "windows" {
				if isPosixShellOnWindows(profile) {
					return generator.ExecPosix
				}
				return generator.ExecCmd
			}
			return generator.ExecPosix
		}
	}
	return host
}

func isPosixShellOnWindows(profile environment.SystemProfile) bool {
	if strings.ToLower(profile.OS) != "windows" {
		return false
	}
	switch strings.ToLower(profile.Shell) {
	case "bash", "zsh", "sh", "fish":
		return true
	default:
		return false
	}
}

func buildCommand(ctx context.Context, gen generator.GeneratedCommand, profile environment.SystemProfile) (*exec.Cmd, error) {
	host := effectiveExecHost(gen, profile)
	return buildCommandForHost(ctx, host, gen, profile)
}

func buildCommandForHost(ctx context.Context, host generator.ExecHost, gen generator.GeneratedCommand, profile environment.SystemProfile) (*exec.Cmd, error) {
	if gen.Chain != nil {
		return buildChainCommand(ctx, host, gen, profile)
	}
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

func buildChainCommand(ctx context.Context, host generator.ExecHost, gen generator.GeneratedCommand, profile environment.SystemProfile) (*exec.Cmd, error) {
	shellName := gen.Shell
	if shellName == "" {
		shellName = profile.Shell
	}
	script, err := BuildValidatedChainScript(shellName, gen.Chain, profile)
	if err != nil {
		return nil, err
	}
	if host == generator.ExecDirect {
		host = chainExecHostForProfile(profile)
	}
	switch host {
	case generator.ExecPowerShell:
		exe, err := ResolvePowerShell(profile)
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, exe, "-NoProfile", "-NonInteractive", "-Command", script), nil
	case generator.ExecCmd:
		exe, err := ResolveCmd()
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, exe, "/c", script), nil
	case generator.ExecPosix:
		exe, err := ResolvePosixShell()
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, exe, "-c", script), nil
	default:
		return buildChainCommand(ctx, chainExecHostForProfile(profile), gen, profile)
	}
}

func chainExecHostForProfile(profile environment.SystemProfile) generator.ExecHost {
	switch strings.ToLower(strings.TrimSpace(profile.Shell)) {
	case "cmd":
		return generator.ExecCmd
	case "powershell", "pwsh":
		return generator.ExecPowerShell
	default:
		if strings.EqualFold(profile.OS, "windows") {
			return generator.ExecPowerShell
		}
		return generator.ExecPosix
	}
}
