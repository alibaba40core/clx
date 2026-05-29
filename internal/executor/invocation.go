package executor

import (
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

// FormatInvocation returns a display string for the effective subprocess invocation.
func FormatInvocation(gen generator.GeneratedCommand, profile environment.SystemProfile) (string, error) {
	host := effectiveExecHost(gen, profile)
	switch host {
	case generator.ExecDirect:
		return QuoteArgv(gen.Shell, gen.Argv), nil
	case generator.ExecPowerShell:
		exe, err := ResolvePowerShell(profile)
		if err != nil {
			return "", err
		}
		script, err := BuildValidatedScript("powershell", gen.Argv)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s -NoProfile -NonInteractive -Command %s",
			QuotePowerShell(exe), QuotePowerShell(script)), nil
	case generator.ExecCmd:
		exe, err := ResolveCmd()
		if err != nil {
			return "", err
		}
		script, err := BuildValidatedScript("cmd", gen.Argv)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s /c %s", QuoteCmd(exe), QuoteCmd(script)), nil
	case generator.ExecPosix:
		exe, err := ResolvePosixShell()
		if err != nil {
			return "", err
		}
		shellName := profile.Shell
		if shellName == "" {
			shellName = "sh"
		}
		script, err := BuildValidatedScript(shellName, gen.Argv)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s -c %s", QuotePOSIX(exe), QuotePOSIX(script)), nil
	default:
		return QuoteArgv(gen.Shell, gen.Argv), nil
	}
}

// shellNameForScript maps ExecHost to the script quoter shell name.
func shellNameForScript(host generator.ExecHost, profile environment.SystemProfile) string {
	switch host {
	case generator.ExecPowerShell:
		return "powershell"
	case generator.ExecCmd:
		return "cmd"
	case generator.ExecPosix:
		if s := strings.ToLower(profile.Shell); s != "" {
			return s
		}
		return "sh"
	default:
		if s := strings.ToLower(profile.Shell); s != "" {
			return s
		}
		return "sh"
	}
}
