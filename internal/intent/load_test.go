package intent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRulesFromModuleRoot(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngineFromModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if len(eng.rules) < 5 {
		t.Fatalf("expected at least 5 rules, got %d", len(eng.rules))
	}
	_ = root
}

func TestLoadRulesFromFS(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	rules, err := LoadRulesFromFS(os.DirFS(root), "internal/builtin/rules")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rules {
		if r.Intent == "find_file" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("find_file not loaded")
	}
	_ = filepath.Join(root, "internal/builtin/rules")
}

func TestParseRulesFileSingleIntentBackCompat(t *testing.T) {
	t.Parallel()
	data := []byte(`intent: foo
examples:
  - foo bar
strategies:
  linux:
    primary: "foo bar"
`)
	rules, err := parseRulesFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].Intent != "foo" {
		t.Fatalf("got %+v", rules)
	}
}

func TestParseRulesFileIntentsList(t *testing.T) {
	t.Parallel()
	data := []byte(`intents:
  - intent: alpha
    examples:
      - alpha
    strategies:
      linux:
        primary: "alpha"
  - intent: beta
    examples:
      - beta {{name}}
    strategies:
      linux:
        primary: "beta {{name}}"
`)
	rules, err := parseRulesFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Intent != "alpha" || rules[1].Intent != "beta" {
		t.Fatalf("intent order/names wrong: %+v", rules)
	}
}

func TestParseRulesFileRejectsBothShapes(t *testing.T) {
	t.Parallel()
	data := []byte(`intent: foo
examples:
  - foo
strategies:
  linux:
    primary: "foo"
intents:
  - intent: bar
    examples:
      - bar
    strategies:
      linux:
        primary: "bar"
`)
	_, err := parseRulesFile(data)
	if err == nil {
		t.Fatal("expected error for mixed shapes")
	}
	if !strings.Contains(err.Error(), "both intent and intents") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRulesFileChainStrategy(t *testing.T) {
	t.Parallel()
	data := []byte(`intent: find_modified_today
examples:
  - files modified today
strategies:
  powershell:
    chain:
      stages:
        - tokens:
            - value: Get-ChildItem
            - value: -Recurse
        - tokens:
            - value: Where-Object
            - value: "{ $_.x }"
              expr: true
      connectors:
        - pipe
`)
	rules, err := parseRulesFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || !rules[0].Strategies["powershell"].HasChain() {
		t.Fatalf("chain not parsed: %+v", rules[0].Strategies)
	}
}

func TestParseRulesFileEmptyIsSkippedByLoader(t *testing.T) {
	t.Parallel()
	data := []byte(`# just a comment, no intents declared
`)
	_, err := parseRulesFile(data)
	if err != errFileNoIntents {
		t.Fatalf("expected errFileNoIntents, got %v", err)
	}
}
