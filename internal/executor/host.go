package executor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/alibaba40core/clx/internal/environment"
)

var (
	ErrNoPowerShell = errors.New("executor: powershell or pwsh not found")
	ErrNoCmd        = errors.New("executor: cmd.exe not found")
	ErrNoPosixShell = errors.New("executor: posix shell not found")
)

var (
	hostCacheMu sync.Mutex
	hostCache   = make(map[string]string)
)

func cachedHost(key string, resolve func() (string, error)) (string, error) {
	hostCacheMu.Lock()
	if v, ok := hostCache[key]; ok {
		hostCacheMu.Unlock()
		return v, nil
	}
	hostCacheMu.Unlock()

	exe, err := resolve()
	if err != nil {
		return "", err
	}

	hostCacheMu.Lock()
	hostCache[key] = exe
	hostCacheMu.Unlock()
	return exe, nil
}

// ResolvePowerShell returns pwsh or powershell.exe based on profile.Shell with LookPath fallbacks.
func ResolvePowerShell(profile environment.SystemProfile) (string, error) {
	key := "ps:" + strings.ToLower(profile.Shell)
	return cachedHost(key, func() (string, error) {
		order := powershellCandidates(profile.Shell)
		for _, name := range order {
			if exe, err := exec.LookPath(name); err == nil {
				return exe, nil
			}
		}
		return "", ErrNoPowerShell
	})
}

func powershellCandidates(shell string) []string {
	switch strings.ToLower(shell) {
	case "pwsh":
		return []string{"pwsh", "powershell", "powershell.exe"}
	case "powershell":
		return []string{"powershell", "powershell.exe", "pwsh"}
	default:
		return []string{"pwsh", "powershell", "powershell.exe"}
	}
}

// ResolveCmd returns cmd.exe via LookPath or %ComSpec%.
func ResolveCmd() (string, error) {
	return cachedHost("cmd", func() (string, error) {
		if exe, err := exec.LookPath("cmd"); err == nil {
			return exe, nil
		}
		if exe, err := exec.LookPath("cmd.exe"); err == nil {
			return exe, nil
		}
		if comspec := os.Getenv("ComSpec"); comspec != "" {
			if _, err := os.Stat(comspec); err == nil {
				return comspec, nil
			}
		}
		return "", ErrNoCmd
	})
}

// ResolvePosixShell returns sh or $SHELL. On Windows, Git Bash / MSYS / Cygwin
// sessions expose bash via $SHELL or PATH.
func ResolvePosixShell() (string, error) {
	key := "posix"
	if runtime.GOOS == "windows" {
		key = "posix-win"
	}
	return cachedHost(key, func() (string, error) {
		if shell := os.Getenv("SHELL"); shell != "" {
			if _, err := os.Stat(shell); err == nil {
				return shell, nil
			}
		}
		for _, name := range []string{"bash", "bash.exe", "sh", "sh.exe"} {
			if exe, err := exec.LookPath(name); err == nil {
				return exe, nil
			}
		}
		if runtime.GOOS == "windows" {
			return "", fmt.Errorf("%w: not available on windows", ErrNoPosixShell)
		}
		return "", ErrNoPosixShell
	})
}
