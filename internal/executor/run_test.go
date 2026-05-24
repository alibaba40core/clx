package executor

import (
	"context"
	"testing"
	"time"

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

func TestRunRejectsBlockedPolicy(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "/"}}
	err := Run(context.Background(), gen,
		WithRisk(risk.RiskAssessment{Level: risk.High}),
		WithPolicy(policy.Result{Allowed: false, Reason: "test"}),
		WithTimeout(time.Second),
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
