package parser

import "strings"

// AliasLookup resolves a first-token alias name to its expansion value.
type AliasLookup interface {
	Lookup(name string) (value string, ok bool)
}

// expandFirstToken replaces the leading token when lookup hits (single-level only).
func expandFirstToken(body string, lookup AliasLookup) string {
	if lookup == nil {
		return body
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return body
	}
	first, rest := splitFirst(body)
	val, ok := lookup.Lookup(first)
	if !ok {
		return body
	}
	val = strings.TrimSpace(val)
	if val == "" {
		return body
	}
	if rest == "" {
		return val
	}
	return val + " " + rest
}

func splitFirst(s string) (first, rest string) {
	s = strings.TrimSpace(s)
	i := strings.IndexByte(s, ' ')
	if i < 0 {
		return s, ""
	}
	return s[:i], strings.TrimSpace(s[i+1:])
}
