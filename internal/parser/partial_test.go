package parser

import (
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func TestIsPartialShellWindowsVsLinux(t *testing.T) {
	t.Parallel()
	win := environment.SystemProfile{OS: "windows", Shell: "powershell"}
	linux := environment.SystemProfile{OS: "linux", Shell: "bash"}

	if !isPartialShell("grep", win) {
		t.Fatal("grep should be partial on windows")
	}
	if isPartialShell("grep", linux) {
		t.Fatal("grep should be shell on linux")
	}
	if !isPartialShell("find", win) {
		t.Fatal("find should be partial on windows")
	}
	if isPartialShell("find", linux) {
		t.Fatal("find should be shell on linux")
	}
}

func TestUnknownCommandNotPartial(t *testing.T) {
	t.Parallel()
	win := environment.SystemProfile{OS: "windows", Shell: "cmd"}
	if isPartialShell("mytool", win) {
		t.Fatal("unknown command should not be partial")
	}
}
