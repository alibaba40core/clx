package intent

import (
	"strings"

	"github.com/alibaba40core/clx/internal/tokenize"
)

// patternToken is one slot in a pre-tokenized example pattern.
type patternToken struct {
	literal string // set when this slot is a fixed word
	param   string // set when this slot is {{name}}
}

type compiledExample struct {
	rule   Rule
	tokens []patternToken
}

// patternIndex groups compiled examples for fast lookup by token count and
// first literal token.
type patternIndex struct {
	byLenFirst map[int]map[string][]compiledExample
	exact      map[string]exactHit
	total      int
}

type exactHit struct {
	intent string
	params map[string]string
}

func tokenKey(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	for i, t := range tokens {
		if i > 0 {
			b.WriteByte('\x00')
		}
		b.WriteString(t)
	}
	return b.String()
}

func compileRules(rules []Rule) patternIndex {
	idx := patternIndex{
		byLenFirst: make(map[int]map[string][]compiledExample),
		exact:      make(map[string]exactHit),
	}
	for _, rule := range rules {
		for _, ex := range rule.Examples {
			compiled, ok := compileExample(rule, ex)
			if !ok {
				continue
			}
			n := len(compiled.tokens)
			bucket := idx.byLenFirst[n]
			if bucket == nil {
				bucket = make(map[string][]compiledExample)
				idx.byLenFirst[n] = bucket
			}
			key := firstLiteralKey(compiled.tokens)
			bucket[key] = append(bucket[key], compiled)
			idx.total++
			if hit, ok := exactFromCompiled(compiled); ok {
				idx.exact[hit.key] = exactHit{intent: hit.intent, params: hit.params}
			}
		}
	}
	return idx
}

type exactKeyHit struct {
	key    string
	intent string
	params map[string]string
}

func exactFromCompiled(comp compiledExample) (exactKeyHit, bool) {
	tokens := make([]string, len(comp.tokens))
	for i, pt := range comp.tokens {
		if pt.param != "" {
			return exactKeyHit{}, false
		}
		tokens[i] = pt.literal
	}
	return exactKeyHit{
		key:    tokenKey(tokens),
		intent: comp.rule.Intent,
		params: map[string]string{},
	}, true
}

func compileExample(rule Rule, pattern string) (compiledExample, bool) {
	patTokens, err := tokenize.Tokenize(pattern)
	if err != nil {
		return compiledExample{}, false
	}
	tokens := make([]patternToken, len(patTokens))
	for i, pt := range patTokens {
		if name, ok := paramName(pt); ok {
			tokens[i] = patternToken{param: name}
			continue
		}
		tokens[i] = patternToken{literal: pt}
	}
	return compiledExample{rule: rule, tokens: tokens}, true
}

func firstLiteralKey(tokens []patternToken) string {
	if len(tokens) == 0 {
		return ""
	}
	if tokens[0].param != "" {
		return ""
	}
	return tokens[0].literal
}

func matchCompiled(comp compiledExample, input []string) (map[string]string, bool) {
	if len(comp.tokens) != len(input) {
		return nil, false
	}
	params := make(map[string]string)
	for i, pt := range comp.tokens {
		if pt.param != "" {
			params[pt.param] = input[i]
			continue
		}
		if pt.literal != input[i] {
			return nil, false
		}
	}
	if len(params) > maxParams {
		return nil, false
	}
	return params, true
}

func (idx patternIndex) candidates(input []string) []compiledExample {
	n := len(input)
	if n == 0 {
		return nil
	}
	bucket := idx.byLenFirst[n]
	if bucket == nil {
		return nil
	}
	first := input[0]
	out := make([]compiledExample, 0, len(bucket[first])+len(bucket[""]))
	out = append(out, bucket[first]...)
	if first != "" {
		out = append(out, bucket[""]...)
	}
	return out
}

func buildIntentMap(rules []Rule) map[string]Rule {
	m := make(map[string]Rule, len(rules))
	for _, r := range rules {
		m[r.Intent] = r
	}
	return m
}
