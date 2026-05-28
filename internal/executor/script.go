package executor

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyScriptArgv = errors.New("executor: empty argv for script")
	ErrScriptMetachar  = errors.New("executor: argv element contains shell metacharacters")
)

// baseScriptMetachars are forbidden in argv elements assembled into host scripts.
// cmdExtraMetachars adds % because cmd.exe expands %VAR% inside quoted /c scripts.
const (
	baseScriptMetachars = ";|&`$()><\n\r"
	cmdExtraMetachars   = "%"
)

func metacharsForShell(shell string) string {
	if strings.EqualFold(shell, "cmd") {
		return baseScriptMetachars + cmdExtraMetachars
	}
	return baseScriptMetachars
}

// BuildValidatedScript joins argv into a single host script using shell quoting.
// Every element is checked for injection metacharacters before quoting.
func BuildValidatedScript(shell string, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", ErrEmptyScriptArgv
	}
	bad := metacharsForShell(shell)
	for _, arg := range argv {
		if arg == "" {
			return "", ErrEmptyScriptArgv
		}
		if strings.ContainsAny(arg, bad) {
			return "", fmt.Errorf("%w in %q", ErrScriptMetachar, arg)
		}
	}
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
	script := strings.Join(parts, " ")
	if script == "" {
		return "", ErrEmptyScriptArgv
	}
	return script, nil
}
