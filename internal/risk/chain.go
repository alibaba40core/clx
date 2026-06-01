package risk

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
)

// assessChain classifies a multi-stage command: max risk over stages, expr scan, Medium floor.
func assessChain(ctx context.Context, gen generator.GeneratedCommand) RiskAssessment {
	var best RiskAssessment
	best.Level = Low
	best.Reason = "read-only or safe seed command"

	for _, st := range gen.Chain.Stages {
		argv := st.PlainArgv()
		stageGen := generator.GeneratedCommand{Argv: argv}
		ra, _ := Assess(ctx, stageGen)
		if ra.Level > best.Level {
			best = ra
		}
		for _, tok := range st.Tokens {
			if !tok.Expr {
				continue
			}
			if exprDestructive(tok.Value) {
				h := high("destructive pattern in chain expression")
				if h.Level > best.Level {
					best = h
				}
			}
		}
	}

	if gen.Chain.HasExpr() && best.Level < Medium {
		return medium("command chain contains expression tokens")
	}
	return best
}

func exprDestructive(text string) bool {
	tokens := strings.Fields(text)
	if len(tokens) == 0 {
		return false
	}
	if _, ok := destructiveArgv[strings.ToLower(tokens[0])]; ok {
		return true
	}
	if destructiveArgvPattern(tokens) {
		return true
	}
	if recursiveDeletePattern(tokens) {
		return true
	}
	lower := strings.ToLower(text)
	for verb := range destructiveArgv {
		if strings.Contains(lower, verb) {
			return true
		}
	}
	return false
}
