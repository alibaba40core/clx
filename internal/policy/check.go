package policy

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

// Check applies block-list policy (Phase 1.6 stub).
func Check(ctx context.Context, gen generator.GeneratedCommand, _ risk.RiskAssessment) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	pol, err := Load(ctx)
	if err != nil {
		return Result{}, err
	}

	hay := strings.ToLower(gen.Command)
	for _, a := range gen.Argv {
		hay += " " + strings.ToLower(a)
	}

	for _, pattern := range pol.Blocked {
		p := strings.ToLower(strings.TrimSpace(pattern))
		if p == "" {
			continue
		}
		if strings.Contains(hay, p) {
			return Result{Allowed: false, Reason: "matches blocked pattern: " + pattern}, nil
		}
	}

	return AllowedResult(), nil
}
