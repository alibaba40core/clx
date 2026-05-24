package tokenize

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	MaxInputBytes = 8 * 1024
	MaxTokens     = 256
)

var (
	ErrInputTooLong  = errors.New("input exceeds maximum length")
	ErrTooManyTokens = errors.New("too many tokens")
)

// Tokenize splits a line into tokens with quote awareness (no shell expansion).
func Tokenize(s string) ([]string, error) {
	if len(s) > MaxInputBytes {
		return nil, ErrInputTooLong
	}

	tokens := make([]string, 0, 32)
	i := 0
	for i < len(s) {
		for i < len(s) && unicode.IsSpace(rune(s[i])) {
			i++
		}
		if i >= len(s) {
			break
		}

		tok, next, err := readToken(s, i)
		if err != nil {
			return nil, err
		}
		i = next
		if tok == "" {
			continue
		}
		tokens = append(tokens, tok)
		if len(tokens) > MaxTokens {
			return nil, ErrTooManyTokens
		}
	}
	return tokens, nil
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
