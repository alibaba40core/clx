package memory

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

// Resolver reuses the last session intent for narrow follow-up inputs.
type Resolver struct {
	store *Store
}

// NewResolver returns a memory-backed intent resolver.
func NewResolver(store *Store) *Resolver {
	return &Resolver{store: store}
}

// Resolve implements intent.Resolver.
func (r *Resolver) Resolve(ctx context.Context, req parser.Request) (intent.ResolvedIntent, error) {
	if err := ctx.Err(); err != nil {
		return intent.ResolvedIntent{}, err
	}
	if r == nil || r.store == nil {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	last, ok := r.store.LastCommand()
	if !ok || last.Intent == "" {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	if !isFollowUp(req, last) {
		return intent.ResolvedIntent{}, intent.ErrNotFound
	}
	params := mergeFollowUpParams(last.Params, req)
	return intent.ResolvedIntent{
		Intent:     last.Intent,
		Params:     params,
		Confidence: 0.75,
		Source:     intent.SourceMemory,
	}, nil
}

func isFollowUp(req parser.Request, last CommandEntry) bool {
	if req.InputType != parser.InputNaturalLanguage {
		return false
	}
	tokens := req.Tokens
	if len(tokens) == 0 {
		return false
	}
	switch strings.ToLower(tokens[0]) {
	case "again", "same", "repeat":
		return true
	}
	if len(tokens) > 6 {
		return false
	}
	return sharesParamTokens(req, last.Params)
}

func sharesParamTokens(req parser.Request, prev map[string]string) bool {
	if len(prev) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(prev)*2)
	for _, v := range prev {
		for _, tok := range strings.Fields(strings.ToLower(v)) {
			if len(tok) >= 3 {
				seen[tok] = struct{}{}
			}
		}
	}
	raw := strings.ToLower(req.RawInput)
	for tok := range seen {
		if strings.Contains(raw, tok) {
			return true
		}
	}
	return false
}

func mergeFollowUpParams(prev map[string]string, req parser.Request) map[string]string {
	out := make(map[string]string, len(prev)+2)
	for k, v := range prev {
		out[k] = v
	}
	// If user supplied a filename-like last token, map to common param keys.
	if len(req.Tokens) >= 1 {
		last := req.Tokens[len(req.Tokens)-1]
		if strings.Contains(last, ".") || !strings.Contains(strings.ToLower(req.RawInput), "again") {
			for _, key := range []string{"file", "filename", "path"} {
				if _, ok := out[key]; ok {
					out[key] = last
					break
				}
			}
			if _, has := out["file"]; !has {
				if _, has = out["filename"]; !has {
					out["file"] = last
				}
			}
		}
	}
	return out
}
