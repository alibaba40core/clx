package yamlutil

import (
	"strings"
	"testing"
)

func TestDecodeList(t *testing.T) {
	t.Parallel()
	yaml := `intent: find_file
examples:
  - locate {{filename}}
  - find file {{filename}}
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := root.GetString("intent"); got != "find_file" {
		t.Fatalf("intent=%q", got)
	}
	list, ok := root.GetStringList("examples")
	if !ok || len(list) != 2 {
		t.Fatalf("examples=%v ok=%v", list, ok)
	}
}

func TestDecodeNestedListMap(t *testing.T) {
	t.Parallel()
	yaml := `intents:
  - intent: list_dir
    examples:
      - ls {{path}}
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	intentsNode, ok := root.GetChild("intents")
	if !ok || len(intentsNode.List) != 1 {
		t.Fatal("expected one intent")
	}
	item := intentsNode.List[0]
	if got, _ := item.GetString("intent"); got != "list_dir" {
		t.Fatalf("intent=%q", got)
	}
	ex, ok := item.GetStringList("examples")
	if !ok || ex[0] != "ls {{path}}" {
		t.Fatalf("examples=%v", ex)
	}
}

func TestDecodeInlineCommentWithPipe(t *testing.T) {
	t.Parallel()
	yaml := `safety:
  mode: medium          # low | medium | high | custom
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := root.GetString("safety", "mode")
	if !ok || got != "medium" {
		t.Fatalf("safety.mode=%q ok=%v", got, ok)
	}
}

func TestDecodeQuotedHashPreserved(t *testing.T) {
	t.Parallel()
	yaml := `value: "quoted#value"
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := root.GetString("value")
	if !ok || got != "quoted#value" {
		t.Fatalf("value=%q ok=%v", got, ok)
	}
}

func TestDecodeUnquotedHashPreserved(t *testing.T) {
	t.Parallel()
	yaml := `value: a#b
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := root.GetString("value")
	if !ok || got != "a#b" {
		t.Fatalf("value=%q ok=%v", got, ok)
	}
}

func TestDecodeFullLineCommentSkipped(t *testing.T) {
	t.Parallel()
	yaml := `# full line comment
provider: ollama
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := root.GetString("provider")
	if !ok || got != "ollama" {
		t.Fatalf("provider=%q ok=%v", got, ok)
	}
}

func TestDecodeInlineCommentBooleanValue(t *testing.T) {
	t.Parallel()
	yaml := `require_confirmation: true   # used when mode=custom
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := root.GetString("require_confirmation")
	if !ok || got != "true" {
		t.Fatalf("require_confirmation=%q ok=%v", got, ok)
	}
}
