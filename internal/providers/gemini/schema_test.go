package gemini

import "testing"

func TestStripAdditionalProperties(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"type":     "object",
		"required": []any{"intent", "params"},
		"additionalProperties": false,
		"properties": map[string]any{
			"intent": map[string]any{
				"type": "string",
				"enum": []any{"find_file", "list_dir"},
			},
			"params": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"filename": map[string]any{"type": "string"},
				},
			},
		},
	}

	out := stripAdditionalProperties(input)

	// Top level should not have additionalProperties.
	if _, ok := out["additionalProperties"]; ok {
		t.Fatal("top-level additionalProperties not stripped")
	}

	// Nested params should not have additionalProperties.
	props := out["properties"].(map[string]any)
	params := props["params"].(map[string]any)
	if _, ok := params["additionalProperties"]; ok {
		t.Fatal("nested additionalProperties not stripped")
	}

	// Other fields should be preserved.
	if out["type"] != "object" {
		t.Fatal("type lost")
	}
	intent := props["intent"].(map[string]any)
	if intent["type"] != "string" {
		t.Fatal("intent.type lost")
	}
}

func TestStripAdditionalPropertiesNil(t *testing.T) {
	t.Parallel()
	out := stripAdditionalProperties(nil)
	if out != nil {
		t.Fatalf("expected nil, got %v", out)
	}
}
