package ollama

import "errors"

// Local sentinel errors for the HTTP client (mapped to providers.* in Provider.ResolveIntent).
var (
	errUnavailable = errors.New("ollama unavailable")
	errTimeout     = errors.New("ollama timeout")
	errInvalidResp = errors.New("ollama invalid response")
	errNoMatch     = errors.New("ollama no match")
)
