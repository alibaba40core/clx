package environment

import (
	"testing"
)

func stubParentProcessBaseName(t *testing.T) {
	t.Helper()
	orig := parentProcessBaseNameFunc
	parentProcessBaseNameFunc = func() string { return "" }
	t.Cleanup(func() { parentProcessBaseNameFunc = orig })
}

func TestDetectShellUnix(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	if got := detectShellUnix(); got != "zsh" {
		t.Fatalf("got %q", got)
	}
}

func TestDetectShellWindowsPowerShell(t *testing.T) {
	stubParentProcessBaseName(t)
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
	stubParentProcessBaseName(t)
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

func TestDetectShellWindowsGitBash(t *testing.T) {
	stubParentProcessBaseName(t)
	t.Setenv("MSYSTEM", "MINGW64")
	t.Setenv("SHELL", `/usr/bin/bash`)
	t.Setenv("POWERSHELL_VERSION", "7.4.0")
	t.Setenv("POWERSHELL_DISTRO_NAME", "")
	t.Setenv("ComSpec", `C:\Windows\System32\cmd.exe`)
	if got := detectShellWindows(); got != "bash" {
		t.Fatalf("got %q want bash (MSYSTEM must beat POWERSHELL_VERSION)", got)
	}
}

func TestDetectShellWindowsParentWins(t *testing.T) {
	cases := []struct {
		name      string
		parentExe string
		want      string
	}{
		{
			name:      "powershell_parent",
			parentExe: "powershell.exe",
			want:      "powershell",
		},
		{
			name:      "cmd_parent",
			parentExe: "cmd.exe",
			want:      "cmd",
		},
		{
			name:      "mintty_parent",
			parentExe: "mintty.exe",
			want:      "bash",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			orig := parentProcessBaseNameFunc
			parentProcessBaseNameFunc = func() string { return tc.parentExe }
			t.Cleanup(func() { parentProcessBaseNameFunc = orig })

			// Conflicting env must not override parent detection.
			t.Setenv("MSYSTEM", "MINGW64")
			t.Setenv("POWERSHELL_VERSION", "7.4.0")
			t.Setenv("POWERSHELL_DISTRO_NAME", "Ubuntu")
			t.Setenv("ComSpec", `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`)

			if got := detectShellWindows(); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestDetectShellVersionFromEnv(t *testing.T) {
	t.Setenv("POWERSHELL_VERSION", "7.4.0")
	if got := detectShellVersion(); got != "7.4.0" {
		t.Fatalf("got %q", got)
	}
}
