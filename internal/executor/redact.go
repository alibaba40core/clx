package executor

import (
	"regexp"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|passwd)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9._-]+`),
	regexp.MustCompile(`(?i)sk-[a-zA-Z0-9]{8,}`),
	regexp.MustCompile(`(?i)-----BEGIN [A-Z ]+-----[\s\S]*?-----END [A-Z ]+-----`),
}

const maxRedactInput = 64 * 1024

// Redact removes likely secrets from s for safe display/logging.
func Redact(s string) string {
	if len(s) > maxRedactInput {
		s = s[:maxRedactInput]
	}
	out := s
	for _, re := range secretPatterns {
		out = re.ReplaceAllString(out, "[REDACTED]")
	}
	return out
}

// ContainsSecret reports whether s matches any secret-shaped pattern used by Redact.
func ContainsSecret(s string) bool {
	if len(s) > maxRedactInput {
		s = s[:maxRedactInput]
	}
	for _, re := range secretPatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
