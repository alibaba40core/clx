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
	// PowerShell sets PSModulePath; pwsh sets POWERSHELL_DISTRO_NAME or version env.
	if os.Getenv("PSModulePath") != "" || os.Getenv("POWERSHELL_VERSION") != "" {
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
	// Git Bash / MSYS often set MSYSTEM.
	if msys := os.Getenv("MSYSTEM"); msys != "" {
		return "bash"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return detectShellUnix()
	}
	return "cmd"
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
