package capabilities

import (
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
)

// defaultKey is the strategy key matched by every profile as last-resort
// fallback. Use it in rules whose argv is identical across OS/shell (e.g.
// portable binaries like git, docker).
const defaultKey = "default"

// candidateKeys returns strategy map keys eligible for the profile OS/shell,
// in descending preference. The defaultKey is always last so OS-specific
// strategies win when present.
func candidateKeys(profile environment.SystemProfile) []string {
	os := strings.ToLower(profile.OS)
	shell := strings.ToLower(profile.Shell)

	var base []string
	switch os {
	case "windows":
		switch shell {
		case "cmd":
			base = []string{"cmd", "windows", "powershell"}
		case "bash", "zsh", "sh", "fish":
			// Git Bash / MSYS: GNU tools (linux) + native Windows ping/netstat (windows).
			base = []string{"windows", "linux", "cmd", "powershell"}
		case "pwsh":
			base = []string{"pwsh", "powershell", "windows", "cmd"}
		case "powershell":
			base = []string{"powershell", "windows", "cmd"}
		case "":
			base = []string{"powershell", "windows", "cmd"}
		default:
			base = []string{"powershell", "cmd", "windows"}
		}
	case "darwin":
		base = []string{"darwin", "linux"}
	case "linux":
		base = []string{"linux", "darwin"}
	default:
		base = []string{os, "linux"}
	}
	return append(base, defaultKey)
}

// strategyMatchesKey reports whether a strategy entry key applies to this profile.
func strategyMatchesKey(strategyKey string, profile environment.SystemProfile) bool {
	key := strings.ToLower(strategyKey)
	candidates := candidateKeys(profile)
	for _, c := range candidates {
		if key == c {
			return true
		}
	}
	os := strings.ToLower(profile.OS)
	if os == "linux" || os == "darwin" {
		return key == os || strings.HasPrefix(key, os+"_")
	}
	if os == "windows" && isPosixShellOnWindows(profile) {
		return key == "linux" || strings.HasPrefix(key, "linux_")
	}
	return false
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

// keyRank orders strategies for tie-breaking: lower rank is preferred.
func keyRank(strategyKey string, profile environment.SystemProfile) int {
	key := strings.ToLower(strategyKey)
	keys := candidateKeys(profile)
	for i, c := range keys {
		if key == c {
			return i
		}
	}
	os := strings.ToLower(profile.OS)
	if (os == "linux" || os == "darwin") && strings.HasPrefix(key, os+"_") {
		return len(keys)
	}
	if os == "windows" && isPosixShellOnWindows(profile) && strings.HasPrefix(key, "linux_") {
		return len(keys)
	}
	return len(keys) + 1
}
