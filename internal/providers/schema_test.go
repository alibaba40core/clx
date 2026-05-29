package providers

import (
	"testing"
)

func TestBuildResponseSchemaIntentEnum(t *testing.T) {
	t.Parallel()
	schema, err := BuildResponseSchema(IntentRequest{
		KnownIntents: []string{"find_file", "list_dir"},
		RuleParams: map[string][]string{
			"find_file": {"filename"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	props := schema["properties"].(map[string]any)
	intent := props["intent"].(map[string]any)
	enum := intent["enum"].([]any)
	if len(enum) != 2 || enum[0] != "find_file" || enum[1] != "list_dir" {
		t.Fatalf("enum = %v", enum)
	}
}

func TestBuildResponseSchemaRejectsEmptyIntents(t *testing.T) {
	t.Parallel()
	_, err := BuildResponseSchema(IntentRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildOpenAIResponseFormat(t *testing.T) {
	t.Parallel()
	schema, err := BuildResponseSchema(IntentRequest{
		KnownIntents: []string{"list_dir"},
	})
	if err != nil {
		t.Fatal(err)
	}
	format := BuildOpenAIResponseFormat(schema)
	if format["type"] != "json_schema" {
		t.Fatalf("format = %v", format)
	}
	js := format["json_schema"].(map[string]any)
	if js["name"] != "clx_intent" || js["strict"] != true {
		t.Fatalf("json_schema = %v", js)
	}
}
