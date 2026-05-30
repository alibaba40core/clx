package executor

import (
	"context"
	"errors"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/policy"
	"github.com/alibaba40core/clx/internal/risk"
)

func TestRunRejectsMissingGates(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"pwd"}}
	err := Run(context.Background(), gen)
	if err != ErrMissingRisk {
		t.Fatalf("got %v", err)
	}

	err = Run(context.Background(), gen, WithRisk(risk.RiskAssessment{Level: risk.Low}))
	if err != ErrMissingPolicy {
		t.Fatalf("got %v", err)
	}
}

func TestRunSurfacesTimeout(t *testing.T) {
	t.Parallel()
	timeout := 200 * time.Millisecond
	var gen generator.GeneratedCommand
	var profile environment.SystemProfile
	if runtime.GOOS == "windows" {
		gen = generator.GeneratedCommand{
			Argv:     []string{"Start-Sleep", "-Seconds", "5"},
			ExecHost: generator.ExecPowerShell,
		}
		profile = environment.SystemProfile{OS: "windows", Shell: "powershell"}
	} else {
		gen = generator.GeneratedCommand{
			Argv:     []string{"sleep", "5"},
			ExecHost: generator.ExecDirect,
		}
		profile = environment.SystemProfile{OS: runtime.GOOS, Shell: "sh"}
	}
	start := time.Now()
	err := Run(context.Background(), gen,
		WithRisk(risk.RiskAssessment{Level: risk.Low}),
		WithPolicy(policy.AllowedResult()),
		WithProfile(profile),
		WithTimeout(timeout),
		WithIO(io.Discard, io.Discard),
	)
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Fatalf("took %v; expected timeout near %v", elapsed, timeout)
	}
	var toe *TimeoutError
	if !errors.As(err, &toe) {
		t.Fatalf("got %v want *TimeoutError", err)
	}
	if !strings.Contains(toe.Error(), "timed out after") {
		t.Fatalf("message %q", toe.Error())
	}
	if toe.After != timeout {
		t.Fatalf("After=%v want %v", toe.After, timeout)
	}
}

func TestRunRejectsMissingProfileForShellHost(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{
		Argv:     []string{"Get-Location"},
		ExecHost: generator.ExecPowerShell,
	}
	err := Run(context.Background(), gen,
		WithRisk(risk.RiskAssessment{Level: risk.Low}),
		WithPolicy(policy.AllowedResult()),
		WithTimeout(time.Second),
	)
	if err != ErrMissingProfile {
		t.Fatalf("got %v want ErrMissingProfile", err)
	}
}

func TestRunRejectsBlockedPolicy(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "/"}}
	err := Run(context.Background(), gen,
		WithRisk(risk.RiskAssessment{Level: risk.High}),
		WithPolicy(policy.Result{Allowed: false, ExecAllowed: false, Reason: "test"}),
		WithTimeout(time.Second),
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
