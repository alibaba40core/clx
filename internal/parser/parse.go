// Package parser normalizes user input into a structured Request.
// Callers pass environment.SystemProfile for profile-aware PartialShell detection.
// Wired into the CLI pipeline in Phase 1.6; consumed by intent in Phase 1.4.
package parser

import (
	"context"
	"errors"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
)

var errEmptyInput = errors.New("empty input")

// Parse normalizes raw user input into a Request.
// When lookup is non-nil, the first token may be expanded via a user alias (single-level).
func Parse(ctx context.Context, raw string, profile environment.SystemProfile, lookup AliasLookup) (Request, error) {
	if err := ctx.Err(); err != nil {
		return Request{}, err
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Request{}, errEmptyInput
	}

	body, hadCLXPrefix := stripCLXPrefix(trimmed)
	body = expandFirstToken(body, lookup)

	if err := ctx.Err(); err != nil {
		return Request{}, err
	}

	tokResult, err := tokenizeInput(body)
	if err != nil {
		return Request{}, err
	}

	if len(tokResult.tokens) == 0 {
		return Request{}, errEmptyInput
	}

	req := Request{
		RawInput:       raw,
		EffectiveInput: body,
		Tokens:         tokResult.tokens,
		Args:           tokResult.args,
	}
	if req.Args == nil {
		req.Args = make(map[string]string)
	}

	if hadCLXPrefix {
		req.InputType = InputCLXInvocation
		return req, nil
	}

	if isNaturalLanguage(body, tokResult.tokens) {
		req.InputType = InputNaturalLanguage
		return req, nil
	}

	if isPartialShell(tokResult.tokens[0], profile) {
		req.InputType = InputPartialShell
		return req, nil
	}

	req.InputType = InputShell
	return req, nil
}

func stripCLXPrefix(s string) (body string, stripped bool) {
	lower := strings.ToLower(s)
	for _, prefix := range []string{"clxmax ", "clx "} {
		if strings.HasPrefix(lower, prefix) {
			return s[len(prefix):], true
		}
	}
	// Exact "clx" or "clxmax" with no trailing body.
	if lower == "clx" || lower == "clxmax" {
		return "", true
	}
	return s, false
}
