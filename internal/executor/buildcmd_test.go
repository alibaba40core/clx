package executor

import (
	"context"
	"os/exec"
	"runtime"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func TestEffectiveExecHostCmdPromotesFindstr(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	if _, err := exec.LookPath("findstr"); err != nil {
		t.Skip("findstr not in PATH")
	}
	gen := generator.GeneratedCommand{
		Argv:     []string{"findstr", "x", "y"},
		ExecHost: generator.ExecCmd,
	}
	if got := effectiveExecHost(gen); got != generator.ExecDirect {
		t.Fatalf("got %v want ExecDirect", got)
	}
}

func TestBuildCommandCmdPromotesFindstr(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	if _, err := exec.LookPath("findstr"); err != nil {
		t.Skip("findstr not in PATH")
	}
	gen := generator.GeneratedCommand{
		Argv:     []string{"findstr", "/C:x", "nul"},
		ExecHost: generator.ExecCmd,
	}
	cmd, err := buildCommand(context.Background(), gen, environment.SystemProfile{OS: "windows", Shell: "cmd"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Path == "" {
		t.Fatal("empty cmd path")
	}
	// Promoted to direct argv: first arg after binary should be findstr or path ending in findstr.
	if len(cmd.Args) < 2 {
		t.Fatalf("args %v", cmd.Args)
	}
	found := false
	for _, a := range cmd.Args {
		if a == "findstr" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected findstr in args %v", cmd.Args)
	}
}
