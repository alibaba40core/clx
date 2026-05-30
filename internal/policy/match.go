package policy

import (
	"strings"
)

// tokenizePattern splits a blocked pattern into argv tokens (whitespace-normalized).
func tokenizePattern(pattern string) []string {
	pattern = strings.TrimSpace(strings.ToLower(pattern))
	if pattern == "" {
		return nil
	}
	fields := strings.Fields(pattern)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Trim(f, `"'`)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// argvMatchesBlocked reports whether pattern tokens appear as an ordered subsequence
// of argv. Single-token patterns require whole-token equality (no substring match).
func argvMatchesBlocked(argv []string, patternTokens []string) bool {
	if len(patternTokens) == 0 || len(argv) == 0 {
		return false
	}
	lower := make([]string, len(argv))
	for i, a := range argv {
		lower[i] = strings.ToLower(a)
	}
	if len(patternTokens) == 1 {
		for _, a := range lower {
			if a == patternTokens[0] {
				return true
			}
		}
		return false
	}
	j := 0
	for _, a := range lower {
		if a == patternTokens[j] {
			j++
			if j == len(patternTokens) {
				return true
			}
		}
	}
	return false
}
