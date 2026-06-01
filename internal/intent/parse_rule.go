package intent

import (
	"fmt"
	"strconv"

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
		child, ok := stratNode.GetChild(k)
		if !ok || child == nil {
			continue
		}
		s := Strategy{}
		if primary, ok := child.GetString("primary"); ok && primary != "" {
			s.Primary = primary
		}
		if argv, ok := child.GetStringList("argv"); ok {
			s.Argv = argv
		}
		if chainNode, ok := child.GetChild("chain"); ok && chainNode != nil {
			chain, err := parseChainSpec(chainNode)
			if err != nil {
				return nil, fmt.Errorf("strategy %q: %w", k, err)
			}
			s.Chain = chain
		}
		if rt, ok := child.GetString("requires_tool"); ok {
			s.RequiresTool = rt
		}
		if pri, ok := child.GetString("priority"); ok && pri != "" {
			if n, err := strconv.Atoi(pri); err == nil {
				s.Priority = n
			}
		}
		if s.Primary == "" && len(s.Argv) == 0 && !s.HasChain() {
			continue
		}
		out[k] = s
	}
	return out, nil
}
