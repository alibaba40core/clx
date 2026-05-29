package generator

import (
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
)

// NewAICommand builds a GeneratedCommand from an AI-provided argv. It selects an
// ExecHost from the model's shell hint (falling back to the active profile) so
// shell builtins (e.g. cmd "dir", "del") run correctly; the executor still
// promotes to direct exec when argv[0] is found on PATH. The argv MUST already
// have passed executor.ValidateGeneratedArgv.
func NewAICommand(argv []string, shellHint, explanation string, profile environment.SystemProfile) GeneratedCommand {
	host := ExecHostForShell(shellHint, profile)
	shell := strings.TrimSpace(shellHint)
	if shell == "" {
		shell = profile.Shell
	}
	return GeneratedCommand{
		Argv:        argv,
		Command:     strings.Join(argv, " "),
		Shell:       shell,
		Explanation: explanation,
		ExecHost:    host,
		AIGenerated: true,
	}
}

// ExecHostForShell maps a shell name to an ExecHost, using the profile as a
// fallback when the hint is empty or unrecognized.
func ExecHostForShell(shell string, profile environment.SystemProfile) ExecHost {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "cmd":
		return ExecCmd
	case "powershell", "pwsh":
		return ExecPowerShell
	case "bash", "sh", "zsh", "fish":
		return ExecPosix
	}
	// Unknown/empty hint: derive from the active profile.
	switch strings.ToLower(strings.TrimSpace(profile.Shell)) {
	case "cmd":
		return ExecCmd
	case "powershell", "pwsh":
		return ExecPowerShell
	case "bash", "sh", "zsh", "fish":
		return ExecPosix
	}
	if strings.EqualFold(profile.OS, "windows") {
		return ExecCmd
	}
	return ExecPosix
}
