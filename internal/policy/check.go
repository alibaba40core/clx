package policy

import (
	"context"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

// Check applies block-list, allow-list, and access-level policy.
func Check(ctx context.Context, gen generator.GeneratedCommand, ra risk.RiskAssessment) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	pol, err := Load(ctx)
	if err != nil {
		return Result{}, err
	}

	if len(gen.Argv) == 0 {
		return AllowedResult(), nil
	}

	for _, pattern := range pol.Blocked {
		tokens := tokenizePattern(pattern)
		if argvMatchesBlocked(gen.Argv, tokens) {
			return Result{Allowed: false, Reason: "matches blocked pattern: " + pattern}, nil
		}
	}

	if len(pol.Allowed) > 0 && !verbOnAllowList(gen.Argv, pol.Allowed) {
		return Result{Allowed: false, Reason: "command verb not on allow list"}, nil
	}

	if ok, reason := accessLevelAllows(ra, pol.AccessLevel); !ok {
		return Result{Allowed: false, Reason: reason}, nil
	}

	return AllowedResult(), nil
}
