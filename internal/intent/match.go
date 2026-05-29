package intent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alibaba40core/clx/internal/tokenize"
)

const maxParams = 16

func matchPattern(pattern string, inputTokens []string) (map[string]string, bool) {
	patTokens, err := tokenize.Tokenize(pattern)
	if err != nil {
		return nil, false
	}
	if len(patTokens) != len(inputTokens) {
		return nil, false
	}

	params := make(map[string]string)
	for i, pt := range patTokens {
		if name, ok := paramName(pt); ok {
			params[name] = inputTokens[i]
			continue
		}
		if pt != inputTokens[i] {
			return nil, false
		}
	}
	if len(params) > maxParams {
		return nil, false
	}
	return params, true
}

func paramName(tok string) (string, bool) {
	if len(tok) < 4 || !strings.HasPrefix(tok, "{{") || !strings.HasSuffix(tok, "}}") {
		return "", false
	}
	name := strings.TrimSpace(tok[2 : len(tok)-2])
	if name == "" {
		return "", false
	}
	return name, true
}

func validateParams(rule Rule, params map[string]string) error {
	if len(rule.Params) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(rule.Params))
	for _, p := range rule.Params {
		allowed[p] = struct{}{}
	}
	for k := range params {
		if _, ok := allowed[k]; !ok {
			return fmt.Errorf("unexpected param %q", k)
		}
	}
	for _, p := range rule.Params {
		if _, ok := params[p]; !ok {
			return fmt.Errorf("missing param %q", p)
		}
	}
	return nil
}

// validateResolvedParams checks params for externally-sourced resolutions (AI, cache).
// When rule.Params is empty, allowed keys are derived from {{placeholders}} in examples.
func validateResolvedParams(rule Rule, params map[string]string) error {
	if err := validateParams(rule, params); err != nil {
		return err
	}
	if len(rule.Params) > 0 {
		return nil
	}
	allowed, err := paramNamesFromExamples(rule)
	if err != nil {
		return err
	}
	for k := range params {
		if _, ok := allowed[k]; !ok {
			return fmt.Errorf("unexpected param %q", k)
		}
	}
	required, err := requiredParamsFromExamples(rule)
	if err != nil {
		return err
	}
	for p := range required {
		if _, ok := params[p]; !ok {
			return fmt.Errorf("missing param %q", p)
		}
	}
	return nil
}

// ExampleParamKeys returns declared rule params, or the union of {{placeholders}} in examples.
func ExampleParamKeys(rule Rule) ([]string, error) {
	if len(rule.Params) > 0 {
		out := append([]string(nil), rule.Params...)
		sort.Strings(out)
		return out, nil
	}
	allowed, err := paramNamesFromExamples(rule)
	if err != nil {
		return nil, err
	}
	if len(allowed) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(allowed))
	for k := range allowed {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

func requiredParamsFromExamples(rule Rule) (map[string]struct{}, error) {
	if len(rule.Examples) == 0 {
		return nil, nil
	}
	var inter map[string]struct{}
	for i, ex := range rule.Examples {
		names, err := paramNamesFromExample(ex)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			inter = names
			continue
		}
		for k := range inter {
			if _, ok := names[k]; !ok {
				delete(inter, k)
			}
		}
	}
	return inter, nil
}

func paramNamesFromExample(ex string) (map[string]struct{}, error) {
	tokens, err := tokenize.Tokenize(ex)
	if err != nil {
		return nil, err
	}
	names := make(map[string]struct{})
	for _, tok := range tokens {
		if name, ok := paramName(tok); ok {
			names[name] = struct{}{}
		}
	}
	return names, nil
}

func paramNamesFromExamples(rule Rule) (map[string]struct{}, error) {
	allowed := make(map[string]struct{})
	for _, ex := range rule.Examples {
		tokens, err := tokenize.Tokenize(ex)
		if err != nil {
			return nil, err
		}
		for _, tok := range tokens {
			if name, ok := paramName(tok); ok {
				allowed[name] = struct{}{}
			}
		}
	}
	return allowed, nil
}
