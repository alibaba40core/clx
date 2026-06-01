package generator

import "github.com/alibaba40core/clx/internal/environment"

const ShellChainExplanation = "chained shell command"

// NewShellChainCommand builds a GeneratedCommand from user shell input with connectors.
func NewShellChainCommand(chain *CommandChain, profile environment.SystemProfile) GeneratedCommand {
	shell := profile.Shell
	return GeneratedCommand{
		Chain:       chain,
		Shell:       shell,
		Explanation: ShellChainExplanation,
		ExecHost:    chainExecHost(shell, profile),
	}
}
