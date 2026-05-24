package capabilities

import (
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
)

// candidateKeys returns strategy map keys eligible for the profile OS/shell.
func candidateKeys(profile environment.SystemProfile) []string {
	os := strings.ToLower(profile.OS)
	shell := strings.ToLower(profile.Shell)

	switch os {
	case "windows":
		switch shell {
		case "cmd":
			return []string{"cmd", "windows", "powershell"}
		case "powershell", "":
			return []string{"powershell", "windows", "cmd"}
		default:
			return []string{"powershell", "cmd", "windows"}
		}
	case "darwin":
		return []string{"darwin", "linux"}
	case "linux":
		return []string{"linux", "darwin"}
	default:
		return []string{os, "linux"}
	}
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
	return false
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
	return len(keys) + 1
}
