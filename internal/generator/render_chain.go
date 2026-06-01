package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/intent"
)

func renderChain(
	ctx context.Context,
	spec *intent.ChainSpec,
	params map[string]string,
) (*CommandChain, error) {
	if spec == nil || len(spec.Stages) < 2 {
		return nil, fmt.Errorf("chain: invalid spec")
	}
	stages := make([]ChainStage, 0, len(spec.Stages))
	for _, stSpec := range spec.Stages {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		toks := stSpec.Tokens
		if len(toks) == 0 && len(stSpec.Argv) > 0 {
			toks = make([]intent.ChainTokenSpec, 0, len(stSpec.Argv))
			for _, v := range stSpec.Argv {
				toks = append(toks, intent.ChainTokenSpec{Value: v})
			}
		}
		stage := ChainStage{Tokens: make([]ChainToken, 0, len(toks))}
		for _, t := range toks {
			val, err := substituteSlot(t.Value, params)
			if err != nil {
				return nil, err
			}
			stage.Tokens = append(stage.Tokens, ChainToken{Value: val, Expr: t.Expr})
		}
		if len(stage.Tokens) == 0 {
			return nil, fmt.Errorf("chain: empty stage after substitution")
		}
		stages = append(stages, stage)
	}
	conns := make([]ChainConnector, len(stages)-1)
	for i := 0; i < len(stages)-1; i++ {
		conns[i] = ChainPipe
		if i < len(spec.Connectors) {
			switch strings.ToLower(strings.TrimSpace(spec.Connectors[i])) {
			case "and", "&&":
				conns[i] = ChainAnd
			default:
				conns[i] = ChainPipe
			}
		}
	}
	return &CommandChain{Stages: stages, Connectors: conns}, nil
}
