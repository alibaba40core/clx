package openai

import "errors"

var (
	errUnavailable = errors.New("openai unavailable")
	errTimeout     = errors.New("openai timeout")
	errInvalidResp = errors.New("openai invalid response")
	errNoMatch     = errors.New("openai no match")
)
