package policy

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

// CheckOptions tunes policy for safety mode and explain-only runs.
type CheckOptions struct {
	SafetyMode  string
	ExplainOnly bool
}

// Check applies block-list, allow-list (high safety only), and access-level policy.
func Check(ctx context.Context, gen generator.GeneratedCommand, ra risk.RiskAssessment, opts CheckOptions) (Result, error) {
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

	explain := opts.ExplainOnly

	for _, pattern := range pol.Blocked {
		tokens := tokenizePattern(pattern)
		if argvMatchesBlocked(gen.Argv, tokens) {
			return denyResult("matches blocked pattern: "+pattern, explain), nil
		}
	}

	if AllowListActive(opts.SafetyMode, pol.Allowed) && !verbOnAllowList(gen.Argv, pol.Allowed) {
		return denyResult("command verb not on allow list", explain), nil
	}

	if ok, reason := accessLevelAllows(ra, pol.AccessLevel); !ok {
		return denyResult(reason, explain), nil
	}

	return AllowedResult(), nil
}

func enforceAllowList(safetyMode string) bool {
	return strings.EqualFold(strings.TrimSpace(safetyMode), "high")
}
