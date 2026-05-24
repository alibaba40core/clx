package parser

import (
	"errors"
	"strings"
	"unicode"

	"github.com/alibaba40core/clx/internal/tokenize"
)

const maxArgs = 32

var errTooManyArgs = errors.New("too many environment assignments")

type tokenizeResult struct {
	tokens []string
	args   map[string]string
}

func tokenizeInput(s string) (tokenizeResult, error) {
	tokens, err := tokenize.Tokenize(s)
	if err != nil {
		return tokenizeResult{}, err
	}

	args := make(map[string]string)
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if key, val, ok := splitAssignment(tok); ok {
			if len(args) >= maxArgs {
				return tokenizeResult{}, errTooManyArgs
			}
			args[key] = val
			continue
		}
		out = append(out, tok)
	}
	return tokenizeResult{tokens: out, args: args}, nil
}

func splitAssignment(tok string) (key, val string, ok bool) {
	idx := strings.IndexByte(tok, '=')
	if idx <= 0 {
		return "", "", false
	}
	key = tok[:idx]
	if !isAssignmentKey(key) {
		return "", "", false
	}
	return key, tok[idx+1:], true
}

func isAssignmentKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
