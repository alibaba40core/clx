package executor

import (
	"runtime"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func TestFormatInvocationMatchesEffectiveHostWindowsBuiltin(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	// Builtin not on PATH: ExecDirect falls back to cmd /c for display and exec.
	gen := generator.GeneratedCommand{
		Argv:     []string{"zzclx-not-in-path-builtin"},
		ExecHost: generator.ExecDirect,
		Shell:    "cmd",
	}
	profile := environment.SystemProfile{OS: "windows", Shell: "cmd"}

	inv, err := FormatInvocation(gen, profile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(inv, "/c") {
		t.Fatalf("expected cmd /c invocation, got %q", inv)
	}

	host := effectiveExecHost(gen)
	if host != generator.ExecCmd {
		t.Fatalf("effective host %v want ExecCmd", host)
	}
}

func TestFormatInvocationParityWithBuildCommandHost(t *testing.T) {
	gen := generator.GeneratedCommand{
		Argv:     []string{"clx-nonexistent-host-parity-test-xyz"},
		ExecHost: generator.ExecDirect,
	}
	want := effectiveExecHost(gen)
	got := effectiveExecHost(gen)
	if want != got {
		t.Fatalf("host mismatch %v vs %v", want, got)
	}
	if runtime.GOOS == "windows" && want != generator.ExecCmd {
		t.Fatalf("on windows missing binary should fall back to ExecCmd, got %v", want)
	}
	if runtime.GOOS != "windows" && want != generator.ExecPosix {
		t.Fatalf("on unix missing binary should fall back to ExecPosix, got %v", want)
	}
}
