package risk

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
)

var lowVerbs = map[string]struct{}{
	"ls": {}, "grep": {}, "rg": {}, "pwd": {}, "df": {}, "du": {}, "find": {}, "fd": {},
	"dir": {}, "cat": {}, "head": {}, "tail": {}, "git": {},
	"get-location": {}, "get-childitem": {}, "select-string": {}, "findstr": {},
}

var destructive = []string{"rm", "shutdown", "format", "del /f", "remove-item", "rmdir"}

// Assess classifies a generated command (Phase 1.6 heuristic stub).
func Assess(ctx context.Context, gen generator.GeneratedCommand) (RiskAssessment, error) {
	if err := ctx.Err(); err != nil {
		return RiskAssessment{}, err
	}

	joined := strings.ToLower(gen.Command)
	for _, d := range destructive {
		if strings.Contains(joined, d) {
			return RiskAssessment{
				Level:                High,
				Reason:               "destructive command pattern",
				RequiresConfirmation: true,
			}, nil
		}
	}
	if strings.Contains(joined, "-rf") || strings.Contains(joined, "/s /q") {
		return RiskAssessment{
			Level:                High,
			Reason:               "recursive or forced delete pattern",
			RequiresConfirmation: true,
		}, nil
	}

	if len(gen.Argv) > 0 {
		verb := strings.ToLower(gen.Argv[0])
		if _, ok := lowVerbs[verb]; ok {
			if verb == "git" && len(gen.Argv) > 1 && strings.ToLower(gen.Argv[1]) != "status" {
				return medium("non-status git command"), nil
			}
			return RiskAssessment{
				Level:                Low,
				Reason:               "read-only or safe seed command",
				RequiresConfirmation: false,
			}, nil
		}
	}

	return medium("unknown command verb"), nil
}

func medium(reason string) RiskAssessment {
	return RiskAssessment{
		Level:                Medium,
		Reason:               reason,
		RequiresConfirmation: true,
	}
}
