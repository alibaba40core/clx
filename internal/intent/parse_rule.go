package intent

import (
	"fmt"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

const maxExamplesPerRule = 64

func parseRuleNode(node *yamlutil.Node) (Rule, error) {
	intent, ok := node.GetString("intent")
	if !ok || intent == "" {
		return Rule{}, fmt.Errorf("rule missing intent")
	}
	examples, ok := node.GetStringList("examples")
	if !ok || len(examples) == 0 {
		return Rule{}, fmt.Errorf("rule %q: missing examples", intent)
	}
	if len(examples) > maxExamplesPerRule {
		return Rule{}, fmt.Errorf("rule %q: too many examples", intent)
	}
	params, _ := node.GetStringList("params")
	strategies, err := parseStrategies(node)
	if err != nil {
		return Rule{}, err
	}
	return Rule{
		Intent:     intent,
		Examples:   examples,
		Params:     params,
		Strategies: strategies,
	}, nil
}

func parseStrategies(node *yamlutil.Node) (map[string]Strategy, error) {
	stratNode, ok := node.GetChild("strategies")
	if !ok || stratNode == nil {
		return nil, nil
	}
	keys, ok := stratNode.GetMapKeys()
	if !ok {
		return nil, nil
	}
	out := make(map[string]Strategy, len(keys))
	for _, k := range keys {
		primary, ok := stratNode.GetString(k, "primary")
		if !ok {
			continue
		}
		out[k] = Strategy{Primary: primary}
	}
	return out, nil
}
