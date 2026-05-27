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
	t.Setenv("ComSpec", `C:\Windows\System32\cmd.exe`)
	t.Setenv("PSModulePath", "C:\\Modules")
	t.Setenv("POWERSHELL_VERSION", "7.4.0")
	t.Setenv("POWERSHELL_DISTRO_NAME", "")
	if got := detectShellWindows(); got != "powershell" {
		t.Fatalf("got %q", got)
	}
}

func TestDetectShellWindowsCmdIgnoresPSModulePath(t *testing.T) {
	t.Setenv("SHELL", "")
	t.Setenv("MSYSTEM", "")
	t.Setenv("PSModulePath", "C:\\Modules")
	t.Setenv("POWERSHELL_VERSION", "")
	t.Setenv("POWERSHELL_DISTRO_NAME", "")
	t.Setenv("ComSpec", `C:\Windows\System32\cmd.exe`)
	got := detectShellWindows()
	if got == "powershell" || got == "pwsh" {
		t.Fatalf("PSModulePath alone must not select PowerShell; got %q", got)
	}
	if got != "cmd" {
		t.Fatalf("got %q want cmd", got)
	}
}

func TestDetectShellVersionFromEnv(t *testing.T) {
	t.Setenv("POWERSHELL_VERSION", "7.4.0")
	if got := detectShellVersion(); got != "7.4.0" {
		t.Fatalf("got %q", got)
	}
}
