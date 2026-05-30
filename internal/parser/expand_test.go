package parser

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

type stubLookup map[string]string

func (s stubLookup) Lookup(name string) (string, bool) {
	v, ok := s[name]
	return v, ok
}

func TestExpandFirstToken(t *testing.T) {
	t.Parallel()
	lookup := stubLookup{"gst": "git status"}
	got := expandFirstToken("gst", lookup)
	if got != "git status" {
		t.Fatalf("got %q", got)
	}
	got = expandFirstToken("gst --short", lookup)
	if got != "git status --short" {
		t.Fatalf("got %q", got)
	}
}

func TestParseWithAliasExpansion(t *testing.T) {
	t.Parallel()
	lookup := stubLookup{"gst": "git status"}
	got, err := Parse(context.Background(), "gst", testProfileLinux(), lookup)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tokens[0] != "git" || got.Tokens[1] != "status" {
		t.Fatalf("tokens=%v", got.Tokens)
	}
}

func TestParseAliasNotChained(t *testing.T) {
	t.Parallel()
	lookup := stubLookup{
		"a": "b",
		"b": "c",
	}
	got, err := Parse(context.Background(), "a", environment.SystemProfile{OS: "linux", Shell: "bash"}, lookup)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tokens[0] != "b" {
		t.Fatalf("expected single expansion to b, got %v", got.Tokens)
	}
}
