package providers

import "log/slog"

// ErrRateLimited indicates the provider rejected the request due to quota or rate limits.
var ErrRateLimited = errRateLimited

// ErrAuth indicates the provider rejected credentials (invalid or missing API key).
var ErrAuth = errAuth

var (
	errRateLimited = errorString("provider rate limited")
	errAuth        = errorString("provider auth failed")
)

type errorString string

func (e errorString) Error() string { return string(e) }

// ClassifyHTTPStatus maps an HTTP status code to a provider sentinel.
// Returns nil for success codes (< 400).
func ClassifyHTTPStatus(status int) error {
	switch {
	case status >= 500:
		return ErrUnavailable
	case status == 429:
		return ErrRateLimited
	case status == 401, status == 403:
		return ErrAuth
	case status >= 400:
		return ErrInvalidResp
	default:
		return nil
	}
}

// HTTPStatusError logs a redacted 4xx/5xx body snippet and returns the classified error.
func HTTPStatusError(status int, provider string, body []byte, logger *slog.Logger) error {
	if status < 400 {
		return nil
	}
	DebugLogHTTPError(logger, provider, status, body)
	return ClassifyHTTPStatus(status)
}
