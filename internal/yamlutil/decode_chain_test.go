package yamlutil

import (
	"strings"
	"testing"
)

func TestDecodeChainStageList(t *testing.T) {
	t.Parallel()
	yaml := `chain:
  stages:
    - tokens:
        - value: Get-ChildItem
        - value: {{path}}
`
	root, err := Decode(strings.NewReader(yaml))
	if err != nil {
		t.Fatal(err)
	}
	chain, ok := root.GetChild("chain")
	if !ok {
		t.Fatal("no chain")
	}
	stages, ok := chain.GetChild("stages")
	if !ok || len(stages.List) == 0 {
		t.Fatalf("stages=%v", stages)
	}
	st0 := stages.List[0]
	toks, ok := st0.GetStringList("tokens", "value")
	if !ok {
		// tokens is list of maps
		tn, _ := st0.GetChild("tokens")
		if tn == nil || len(tn.List) == 0 {
			t.Fatalf("tokens node=%v", tn)
		}
	}
	_ = toks
}
