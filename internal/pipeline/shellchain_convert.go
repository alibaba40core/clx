package pipeline

import (
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/parser"
)

func generatorChainFromShell(sc *parser.ShellChain) *generator.CommandChain {
	if sc == nil || len(sc.Stages) < 2 {
		return nil
	}
	stages := make([]generator.ChainStage, len(sc.Stages))
	for i, argv := range sc.Stages {
		toks := make([]generator.ChainToken, 0, len(argv))
		for _, v := range argv {
			toks = append(toks, generator.ChainToken{
				Value: v,
				Expr:  shellTokenLooksLikeExpr(v),
			})
		}
		stages[i] = generator.ChainStage{Tokens: toks}
	}
	conns := make([]generator.ChainConnector, len(sc.Stages)-1)
	for i := 0; i < len(sc.Stages)-1; i++ {
		conns[i] = generator.ChainPipe
		if i < len(sc.Connectors) {
			switch strings.ToLower(strings.TrimSpace(sc.Connectors[i])) {
			case "and", "&&":
				conns[i] = generator.ChainAnd
			}
		}
	}
	return &generator.CommandChain{Stages: stages, Connectors: conns}
}

func shellTokenLooksLikeExpr(v string) bool {
	if strings.HasPrefix(strings.TrimSpace(v), "{") {
		return true
	}
	return strings.ContainsRune(v, '$')
}
