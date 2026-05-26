package intent

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/parser"
)

func testEngine(t *testing.T) *Engine {
	t.Helper()
	eng, err := NewEngineFromModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

func TestResolveFindFile(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"locate", "help.txt"}}
	got, err := eng.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "find_file" || got.Params["filename"] != "help.txt" || got.Source != SourceRule {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveSearchText(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"grep", "errors", "logs.txt"}}
	got, err := eng.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "search_text_in_file" {
		t.Fatalf("got %+v", got)
	}
	if got.Params["pattern"] != "errors" || got.Params["file"] != "logs.txt" {
		t.Fatalf("params %v", got.Params)
	}
}

func TestResolveListDir(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"ls", "/tmp"}}
	got, err := eng.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "list_dir" || got.Params["path"] != "/tmp" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveCurrentDir(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"pwd"}}
	got, err := eng.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "current_dir" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveDiskUsage(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"disk", "usage"}}
	got, err := eng.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "disk_usage" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveNotFound(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"unknown", "command", "xyz"}}
	_, err := eng.Resolve(context.Background(), req)
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestResolveGitIntents(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	cases := []struct {
		name       string
		tokens     []string
		wantIntent string
		wantParams map[string]string
	}{
		{"git_status", []string{"git", "status"}, "git_status", map[string]string{}},
		{"git_log_bare", []string{"git", "log"}, "git_log", map[string]string{}},
		{"git_log_n", []string{"git", "log", "-n", "50"}, "git_log", map[string]string{"n": "50"}},
		{"git_diff_bare", []string{"git", "diff"}, "git_diff", map[string]string{}},
		{"git_diff_path", []string{"git", "diff", "main.go"}, "git_diff_path", map[string]string{"path": "main.go"}},
		{"git_branch_list", []string{"git", "branch"}, "git_branch_list", map[string]string{}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := eng.Resolve(context.Background(), parser.Request{Tokens: tc.tokens})
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if got.Intent != tc.wantIntent {
				t.Fatalf("intent: got %q want %q", got.Intent, tc.wantIntent)
			}
			if got.Source != SourceRule {
				t.Fatalf("source: got %v want SourceRule", got.Source)
			}
			for k, v := range tc.wantParams {
				if got.Params[k] != v {
					t.Fatalf("param %q: got %q want %q", k, got.Params[k], v)
				}
			}
		})
	}
}
