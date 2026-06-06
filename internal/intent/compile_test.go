package intent

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/parser"
)

func TestCompiledExampleMatchesMatchPattern(t *testing.T) {
	cases := []struct {
		pattern string
		input   []string
		want    bool
	}{
		{"locate {{filename}}", []string{"locate", "help.txt"}, true},
		{"grep {{pattern}} {{file}}", []string{"grep", "errors", "logs.txt"}, true},
		{"pwd", []string{"pwd"}, true},
		{"pwd", []string{"cwd"}, false},
		{"find {{filename}}", []string{"locate", "help.txt"}, false},
	}
	for _, tc := range cases {
		comp, ok := compileExample(Rule{Intent: "test"}, tc.pattern)
		if !ok {
			t.Fatalf("compile %q", tc.pattern)
		}
		_, got := matchCompiled(comp, tc.input)
		if got != tc.want {
			t.Fatalf("pattern %q input %v: got %v want %v", tc.pattern, tc.input, got, tc.want)
		}
	}
}

func TestExactMatchMap(t *testing.T) {
	eng, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	got, err := eng.Resolve(context.Background(), parser.Request{Tokens: []string{"pwd"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "current_dir" {
		t.Fatalf("got %q", got.Intent)
	}
	if len(eng.index.exact) == 0 {
		t.Fatal("expected exact-match entries")
	}
}

func TestPatternIndexCandidates(t *testing.T) {
	eng, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	req := parser.Request{Tokens: []string{"pwd"}}
	cands := eng.index.candidates(req.Tokens)
	if len(cands) == 0 {
		t.Fatal("expected pwd candidates")
	}
	found := false
	for _, c := range cands {
		if c.rule.Intent == "current_dir" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("current_dir not in pwd bucket")
	}
}
