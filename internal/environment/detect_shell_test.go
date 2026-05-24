package environment

import (
	"testing"
)

func TestDetectShellUnix(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	if got := detectShellUnix(); got != "zsh" {
		t.Fatalf("got %q", got)
	}
}

func TestDetectShellWindowsPowerShell(t *testing.T) {
	t.Setenv("SHELL", "")
	t.Setenv("ComSpec", "")
	t.Setenv("PSModulePath", "C:\\Modules")
	if got := detectShellWindows(); got != "powershell" {
		t.Fatalf("got %q", got)
	}
}

func TestDetectShellVersionFromEnv(t *testing.T) {
	t.Setenv("POWERSHELL_VERSION", "7.4.0")
	if got := detectShellVersion(); got != "7.4.0" {
		t.Fatalf("got %q", got)
	}
}
