package executor

import (
	"runtime"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func TestResolvePowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	t.Parallel()
	exe, err := ResolvePowerShell(environment.SystemProfile{Shell: "powershell"})
	if err != nil {
		t.Fatal(err)
	}
	if exe == "" {
		t.Fatal("empty exe")
	}
}

func TestResolveCmd(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	t.Parallel()
	exe, err := ResolveCmd()
	if err != nil {
		t.Fatal(err)
	}
	if exe == "" {
		t.Fatal("empty exe")
	}
}

func TestResolvePosixShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix only")
	}
	t.Parallel()
	exe, err := ResolvePosixShell()
	if err != nil {
		t.Fatal(err)
	}
	if exe == "" {
		t.Fatal("empty exe")
	}
}
