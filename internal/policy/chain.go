package policy

import (
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

func checkChainPolicy(gen generator.GeneratedCommand, ra risk.RiskAssessment, pol File, opts CheckOptions) (Result, error) {
	explain := opts.ExplainOnly

	for _, st := range gen.Chain.Stages {
		argv := st.PlainArgv()
		if len(argv) == 0 {
			continue
		}
		for _, pattern := range pol.Blocked {
			tokens := tokenizePattern(pattern)
			if argvMatchesBlocked(argv, tokens) {
				return denyResult("matches blocked pattern: "+pattern, explain), nil
			}
		}
		for _, tok := range st.Tokens {
			if !tok.Expr {
				continue
			}
			exprArgv := strings.Fields(tok.Value)
			for _, pattern := range pol.Blocked {
				pt := tokenizePattern(pattern)
				if argvMatchesBlocked(exprArgv, pt) {
					return denyResult("matches blocked pattern in expression: "+pattern, explain), nil
				}
			}
		}
	}

	if AllowListActive(opts.SafetyMode, pol.Allowed) {
		for _, st := range gen.Chain.Stages {
			argv := st.PlainArgv()
			if len(argv) == 0 {
				continue
			}
			if !verbOnAllowList(argv, pol.Allowed) {
				return denyResult("command verb not on allow list", explain), nil
			}
		}
	}

	if ok, reason := accessLevelAllows(ra, pol.AccessLevel); !ok {
		return denyResult(reason, explain), nil
	}

	return AllowedResult(), nil
}
