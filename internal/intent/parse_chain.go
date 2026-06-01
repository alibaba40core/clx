package intent

import (
	"fmt"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

func parseChainSpec(node *yamlutil.Node) (*ChainSpec, error) {
	if node == nil {
		return nil, fmt.Errorf("chain: missing node")
	}
	stagesNode, ok := node.GetChild("stages")
	if !ok || stagesNode == nil || len(stagesNode.List) == 0 {
		return nil, fmt.Errorf("chain: missing stages")
	}
	stages := make([]ChainStageSpec, 0, len(stagesNode.List))
	for i, item := range stagesNode.List {
		st, err := parseChainStage(item)
		if err != nil {
			return nil, fmt.Errorf("chain stage %d: %w", i, err)
		}
		stages = append(stages, st)
	}
	if len(stages) < 2 {
		return nil, fmt.Errorf("chain: need at least 2 stages")
	}
	connectors, _ := node.GetStringList("connectors")
	return &ChainSpec{Stages: stages, Connectors: connectors}, nil
}

func parseChainStage(item *yamlutil.Node) (ChainStageSpec, error) {
	if item == nil {
		return ChainStageSpec{}, fmt.Errorf("empty stage")
	}
	if argv, ok := item.GetStringList("argv"); ok && len(argv) > 0 {
		toks := make([]ChainTokenSpec, 0, len(argv))
		for _, v := range argv {
			toks = append(toks, ChainTokenSpec{Value: v})
		}
		return ChainStageSpec{Tokens: toks}, nil
	}
	tokensNode, ok := item.GetChild("tokens")
	if !ok || tokensNode == nil || len(tokensNode.List) == 0 {
		return ChainStageSpec{}, fmt.Errorf("stage needs argv or tokens")
	}
	toks := make([]ChainTokenSpec, 0, len(tokensNode.List))
	for _, tokNode := range tokensNode.List {
		if tokNode.Scalar != "" {
			toks = append(toks, ChainTokenSpec{Value: tokNode.Scalar})
			continue
		}
		val, ok := tokNode.GetString("value")
		if !ok || val == "" {
			return ChainStageSpec{}, fmt.Errorf("token missing value")
		}
		expr := false
		if es, ok := tokNode.GetString("expr"); ok {
			expr = es == "true" || es == "1"
		}
		toks = append(toks, ChainTokenSpec{Value: val, Expr: expr})
	}
	return ChainStageSpec{Tokens: toks}, nil
}
