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
