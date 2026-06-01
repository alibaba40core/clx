package generator

import (
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
)

// NewCommandFromChain builds a GeneratedCommand from a validated CommandChain.
// script is the display/exec script from executor.BuildValidatedChainScript.
func NewCommandFromChain(chain *CommandChain, shellHint, explanation, script string, profile environment.SystemProfile) GeneratedCommand {
	shell := strings.TrimSpace(shellHint)
	if shell == "" {
		shell = profile.Shell
	}
	host := ExecHostForShell(shellHint, profile)
	if host == ExecDirect {
		host = chainExecHost(shell, profile)
	}
	return GeneratedCommand{
		Chain:       chain,
		Command:     script,
		Shell:       shell,
		Explanation: explanation,
		ExecHost:    host,
		AIGenerated: true,
	}
}

// chainExecHost picks a shell host for multi-stage chains (never ExecDirect).
func chainExecHost(shell string, profile environment.SystemProfile) ExecHost {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "cmd":
		return ExecCmd
	case "powershell", "pwsh":
		return ExecPowerShell
	case "bash", "sh", "zsh", "fish":
		return ExecPosix
	}
	return ExecHostForShell(shell, profile)
}
