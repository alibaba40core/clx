package environment

import "strings"

// shellFromParentExecutable maps a parent executable base name to a shell id, or
// "" if unknown. The mapping is pure string logic and platform-independent; only
// resolving the parent process name (parentProcessBaseName) is OS-specific. Keeping
// this here lets cross-platform tests exercise the mapping directly.
func shellFromParentExecutable(base string) string {
	switch {
	case strings.Contains(base, "pwsh"):
		return "pwsh"
	case strings.Contains(base, "powershell"):
		return "powershell"
	case base == "cmd.exe":
		return "cmd"
	case strings.Contains(base, "mintty"), base == "winpty-agent.exe":
		return "bash"
	case strings.Contains(base, "bash"), strings.Contains(base, "sh"):
		return detectShellUnix()
	default:
		return ""
	}
}
