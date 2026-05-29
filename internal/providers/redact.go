package providers

import (
	"log/slog"

	"github.com/alibaba40core/clx/internal/executor"
)

const maxLogSnippet = 256

// RedactHTTPBody returns a redacted, length-capped snippet safe for debug logs (C7).
func RedactHTTPBody(body []byte) string {
	s := executor.Redact(string(body))
	if len(s) > maxLogSnippet {
		return s[:maxLogSnippet] + "…"
	}
	return s
}

// RedactError returns a redacted error string safe for slog fields (C7).
func RedactError(err error) string {
	if err == nil {
		return ""
	}
	return executor.Redact(err.Error())
}

// DebugLogHTTPError logs a redacted HTTP error body snippet when logger is non-nil.
func DebugLogHTTPError(logger *slog.Logger, provider string, status int, body []byte) {
	if logger == nil || len(body) == 0 {
		return
	}
	logger.Debug("provider http error",
		"provider", provider,
		"status", status,
		"body_snippet", RedactHTTPBody(body),
	)
}
