package generator

import "strings"

// ExecHost selects how the executor runs GeneratedCommand.Argv.
type ExecHost int

const (
	// ExecDirect runs argv[0] as a PATH binary with remaining args (no shell).
	ExecDirect ExecHost = iota
	// ExecPowerShell runs a validated script via pwsh/powershell.exe -Command.
	ExecPowerShell
	// ExecCmd runs a validated script via cmd.exe /c.
	ExecCmd
	// ExecPosix runs a validated script via sh -c (Unix fallback).
	ExecPosix
)

// inferExecHost picks the execution host from the selected strategy key only.
// PATH-based promotion (cmd strategy + findstr on PATH) is resolved in the executor.
func inferExecHost(strategyKey string) ExecHost {
	switch strings.ToLower(strategyKey) {
	case "powershell", "pwsh":
		return ExecPowerShell
	case "cmd":
		return ExecCmd
	default:
		return ExecDirect
	}
}
