package providers

import (
	"fmt"
	"sort"
)

// BuildResponseSchema returns a JSON Schema object for Ollama's format field.
// Intent names are constrained to req.KnownIntents; param keys are the union of
// all declared rule params (per-intent validation still happens in ValidateResolved).
func BuildResponseSchema(req IntentRequest) (map[string]any, error) {
	if len(req.KnownIntents) == 0 {
		return nil, fmt.Errorf("schema: no known intents")
	}
	if len(req.KnownIntents) > maxKnownIntents {
		return nil, fmt.Errorf("schema: too many known intents: %d", len(req.KnownIntents))
	}

	intents := append([]string(nil), req.KnownIntents...)
	sort.Strings(intents)

	paramProps := unionParamProperties(req.RuleParams)
	paramsSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
	}
	if len(paramProps) > 0 {
		paramsSchema["properties"] = paramProps
	}

	return map[string]any{
		"type":     "object",
		"required": []any{"intent", "params"},
		"properties": map[string]any{
			"intent": map[string]any{
				"type": "string",
				"enum": stringSliceToAny(intents),
			},
			"params": paramsSchema,
			"confidence": map[string]any{
				"type":    "number",
				"minimum": 0,
				"maximum": 1,
			},
		},
	}, nil
}

func unionParamProperties(ruleParams map[string][]string) map[string]any {
	if len(ruleParams) == 0 {
		return nil
	}
	names := make(map[string]struct{})
	for _, ps := range ruleParams {
		for _, p := range ps {
			if p != "" {
				names[p] = struct{}{}
			}
		}
	}
	if len(names) == 0 {
		return nil
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	props := make(map[string]any, len(sorted))
	for _, n := range sorted {
		props[n] = map[string]any{"type": "string"}
	}
	return props
}

func stringSliceToAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}
