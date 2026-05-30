package gemini

// stripAdditionalProperties recursively removes "additionalProperties" keys
// from a JSON Schema map. Gemini's responseSchema uses an OpenAPI 3.0 subset
// that does not support this field; including it causes 400 errors.
func stripAdditionalProperties(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	out := make(map[string]any, len(schema))
	for k, v := range schema {
		if k == "additionalProperties" {
			continue
		}
		switch val := v.(type) {
		case map[string]any:
			out[k] = stripAdditionalProperties(val)
		case []any:
			out[k] = stripSlice(val)
		default:
			out[k] = v
		}
	}
	return out
}

func stripSlice(arr []any) []any {
	out := make([]any, len(arr))
	for i, v := range arr {
		if m, ok := v.(map[string]any); ok {
			out[i] = stripAdditionalProperties(m)
		} else {
			out[i] = v
		}
	}
	return out
}
