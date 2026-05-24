package parser

import (
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
)

// unixCentricCommands lists commands commonly unavailable or aliased on Windows shells.
var unixCentricCommands = map[string]struct{}{
	"locate": {}, "grep": {}, "ls": {}, "which": {}, "uname": {},
	"find": {}, "awk": {}, "sed": {}, "xargs": {}, "wc": {},
	"touch": {}, "ln": {}, "df": {}, "du": {},
}

func isPartialShell(cmd string, profile environment.SystemProfile) bool {
	cmd = strings.ToLower(cmd)
	if _, ok := unixCentricCommands[cmd]; !ok {
		return false
	}
	return isWindowsShell(profile)
}

func isWindowsShell(profile environment.SystemProfile) bool {
	if profile.OS != "windows" {
		return false
	}
	shell := strings.ToLower(profile.Shell)
	return shell == "powershell" || shell == "cmd" || shell == "unknown" || shell == ""
}
