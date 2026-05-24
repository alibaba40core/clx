package intent

import (
	"fmt"
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
