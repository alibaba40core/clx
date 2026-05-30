package gemini

import "errors"

var (
	errUnavailable = errors.New("gemini unavailable")
	errTimeout     = errors.New("gemini timeout")
	errInvalidResp = errors.New("gemini invalid response")
	errNoMatch     = errors.New("gemini no match")
)
