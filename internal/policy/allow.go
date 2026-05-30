package policy

import "strings"

// verbOnAllowList reports whether argv[0] matches an allowed verb (case-insensitive).
func verbOnAllowList(argv []string, allowed []string) bool {
	if len(argv) == 0 {
		return false
	}
	verb := strings.ToLower(strings.TrimSpace(argv[0]))
	for _, a := range allowed {
		if strings.ToLower(strings.TrimSpace(a)) == verb {
			return true
		}
	}
	return false
}
