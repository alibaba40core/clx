package shellchain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

// Symbol returns the native shell operator for a logical connector.
func Symbol(conn generator.ChainConnector, profile environment.SystemProfile) (string, error) {
	shell := strings.ToLower(strings.TrimSpace(profile.Shell))
	switch conn {
	case generator.ChainPipe:
		return "|", nil
	case generator.ChainAnd:
		switch shell {
		case "cmd":
			return "&", nil
		case "powershell", "pwsh":
			if powershellSupportsAnd(profile.ShellVersion) {
				return "&&", nil
			}
			return ";", nil
		default:
			return "&&", nil
		}
	default:
		return "", fmt.Errorf("shellchain: unknown connector %d", conn)
	}
}

// powershellSupportsAnd reports whether the profile shell supports && (PS 7+).
func powershellSupportsAnd(version string) bool {
	version = strings.TrimSpace(strings.ToLower(version))
	if version == "" {
		return false
	}
	if strings.HasPrefix(version, "7.") || strings.HasPrefix(version, "7 ") {
		return true
	}
	if idx := strings.IndexByte(version, '.'); idx > 0 {
		major, err := strconv.Atoi(version[:idx])
		if err == nil && major >= 7 {
			return true
		}
	}
	return strings.Contains(version, "pwsh") || strings.Contains(version, "core")
}
