package environment

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func detectShell() string {
	if runtime.GOOS == "windows" {
		return detectShellWindows()
	}
	return detectShellUnix()
}

func detectShellUnix() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}
	base := strings.ToLower(filepath.Base(shell))
	switch {
	case strings.Contains(base, "zsh"):
		return "zsh"
	case strings.Contains(base, "bash"):
		return "bash"
	case strings.Contains(base, "fish"):
		return "fish"
	case strings.Contains(base, "sh"):
		return "sh"
	default:
		return base
	}
}

func detectShellWindows() string {
	// Prefer the shell that launched this process (cmd.exe vs powershell.exe).
	if base := parentProcessBaseName(); base != "" {
		if shell := shellFromParentExecutable(base); shell != "" {
			return shell
		}
	}
	// Git Bash / MSYS / Cygwin (mintty): use POSIX shell before PowerShell session env.
	if shell := posixShellOnWindowsFromEnv(); shell != "" {
		return shell
	}
	// Session markers set by PowerShell/pwsh (not user-level PSModulePath alone).
	if os.Getenv("POWERSHELL_DISTRO_NAME") != "" {
		return "pwsh"
	}
	if os.Getenv("POWERSHELL_VERSION") != "" {
		return "powershell"
	}
	comspec := os.Getenv("ComSpec")
	if comspec != "" {
		base := strings.ToLower(filepath.Base(comspec))
		if strings.Contains(base, "powershell") || strings.Contains(base, "pwsh") {
			return "powershell"
		}
		if strings.Contains(base, "cmd") {
			return "cmd"
		}
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return detectShellUnix()
	}
	return "cmd"
}

// posixShellOnWindowsFromEnv detects Git Bash, MSYS2, or Cygwin sessions on Windows.
func posixShellOnWindowsFromEnv() string {
	if os.Getenv("MSYSTEM") != "" {
		return "bash"
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	lower := strings.ToLower(shell)
	if strings.Contains(lower, "bash") || strings.Contains(lower, "/git-") {
		return detectShellUnix()
	}
	return ""
}

func detectShellVersion() string {
	if v := os.Getenv("POWERSHELL_VERSION"); v != "" {
		return v
	}
	if v := os.Getenv("BASH_VERSION"); v != "" {
		return v
	}
	if v := os.Getenv("ZSH_VERSION"); v != "" {
		return v
	}
	return ""
}
