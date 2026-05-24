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
