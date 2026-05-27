package generator

import (
	"os/exec"
	"strings"
)

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

// inferExecHost picks the execution host from the selected strategy key and argv.
func inferExecHost(strategyKey string, argv []string) ExecHost {
	key := strings.ToLower(strategyKey)
	switch key {
	case "powershell":
		return ExecPowerShell
	case "cmd":
		if len(argv) > 0 {
			if _, err := exec.LookPath(argv[0]); err == nil {
				return ExecDirect
			}
		}
		return ExecCmd
	default:
		return ExecDirect
	}
}
