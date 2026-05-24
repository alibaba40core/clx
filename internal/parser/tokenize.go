package parser

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	maxInputBytes = 8 * 1024
	maxTokens     = 256
	maxArgs       = 32
)

var (
	errInputTooLong = errors.New("input exceeds maximum length")
	errTooManyTokens = errors.New("too many tokens")
	errTooManyArgs   = errors.New("too many environment assignments")
)

type tokenizeResult struct {
	tokens []string
	args   map[string]string
}

func tokenizeInput(s string) (tokenizeResult, error) {
	if len(s) > maxInputBytes {
		return tokenizeResult{}, errInputTooLong
	}

	tokens := make([]string, 0, 32)
	args := make(map[string]string)
	i := 0
	for i < len(s) {
		// Skip whitespace.
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			i++
		}
		if i >= len(s) {
			break
		}

		var tok string
		var err error
		tok, i, err = readToken(s, i)
		if err != nil {
			return tokenizeResult{}, err
		}
		if tok == "" {
			continue
		}

		if key, val, ok := splitAssignment(tok); ok {
			if len(args) >= maxArgs {
				return tokenizeResult{}, errTooManyArgs
			}
			args[key] = val
			continue
		}

		tokens = append(tokens, tok)
		if len(tokens) > maxTokens {
			return tokenizeResult{}, errTooManyTokens
		}
	}

	return tokenizeResult{tokens: tokens, args: args}, nil
}

func readToken(s string, i int) (string, int, error) {
	if i >= len(s) {
		return "", i, nil
	}

	switch s[i] {
	case '\'':
		return readSingleQuoted(s, i+1)
	case '"':
		return readDoubleQuoted(s, i+1)
	default:
		return readBareToken(s, i)
	}
}

func readSingleQuoted(s string, i int) (string, int, error) {
	start := i
	for i < len(s) {
		if s[i] == '\'' {
			return s[start:i], i + 1, nil
		}
		i++
	}
	return "", len(s), fmt.Errorf("unterminated single-quoted string")
}

func readDoubleQuoted(s string, i int) (string, int, error) {
	var b strings.Builder
	for i < len(s) {
		switch s[i] {
		case '"':
			return b.String(), i + 1, nil
		case '\\':
			if i+1 < len(s) {
				i++
				b.WriteByte(s[i])
				i++
				continue
			}
			return "", len(s), fmt.Errorf("unterminated double-quoted string")
		default:
			b.WriteByte(s[i])
			i++
		}
	}
	return "", len(s), fmt.Errorf("unterminated double-quoted string")
}

func readBareToken(s string, i int) (string, int, error) {
	start := i
	for i < len(s) && !unicode.IsSpace(rune(s[i])) {
		i++
	}
	return s[start:i], i, nil
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
