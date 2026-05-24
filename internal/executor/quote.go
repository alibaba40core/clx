package executor

import (
	"strings"
)

// QuotePOSIX quotes s for POSIX shell display.
func QuotePOSIX(s string) string {
	if s == "" {
		return "''"
	}
	if isSafePOSIX(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func isSafePOSIX(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			continue
		}
		switch c {
		case '-', '_', '.', '/', ':', '@', '+', ',':
			continue
		default:
			return false
		}
	}
	return true
}

// QuotePowerShell quotes s for PowerShell display.
func QuotePowerShell(s string) string {
	if s == "" {
		return "''"
	}
	if isSafePOSIX(s) {
		return s
	}
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

// QuoteCmd quotes s for cmd.exe display.
func QuoteCmd(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\n\v\f\"&|<>^") {
		return s
	}
	escaped := strings.ReplaceAll(s, `"`, `""`)
	return `"` + escaped + `"`
}

// QuoteArgv formats argv for display using the target shell name.
func QuoteArgv(shell string, argv []string) string {
	quote := QuotePOSIX
	switch strings.ToLower(shell) {
	case "powershell", "pwsh":
		quote = QuotePowerShell
	case "cmd":
		quote = QuoteCmd
	}
	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = quote(a)
	}
	return strings.Join(parts, " ")
}
