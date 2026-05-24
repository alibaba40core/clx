package generator

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/capabilities"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

func TestRenderArgvPrimary(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{
		Intent: "search_text_in_file",
		Params: map[string]string{"pattern": "errors", "file": "logs.txt"},
	}
	selected := capabilities.SelectedStrategy{
		Key: "powershell",
		Strategy: intent.Strategy{
			Argv: []string{"Select-String", "{{pattern}}", "{{file}}"},
		},
	}
	profile := environment.SystemProfile{Shell: "powershell"}
	got, err := Render(context.Background(), resolved, selected, profile)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Select-String", "errors", "logs.txt"}
	if len(got.Argv) != len(want) {
		t.Fatalf("argv %v", got.Argv)
	}
	for i := range want {
		if got.Argv[i] != want[i] {
			t.Fatalf("argv[%d]=%q want %q", i, got.Argv[i], want[i])
		}
	}
	if got.Command != "Select-String errors logs.txt" {
		t.Fatalf("command %q", got.Command)
	}
}

func TestRenderTokenizedPrimary(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{
		Intent: "search_text_in_file",
		Params: map[string]string{"pattern": "errors", "file": "logs.txt"},
	}
	selected := capabilities.SelectedStrategy{
		Key:      "linux",
		Strategy: intent.Strategy{Primary: "grep {{pattern}} {{file}}"},
	}
	profile := environment.SystemProfile{Shell: "bash"}
	got, err := Render(context.Background(), resolved, selected, profile)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"grep", "errors", "logs.txt"}
	for i := range want {
		if got.Argv[i] != want[i] {
			t.Fatalf("argv %v", got.Argv)
		}
	}
}

func TestRenderDiskUsageDefaultPath(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{Intent: "disk_usage", Params: map[string]string{}}
	selected := capabilities.SelectedStrategy{
		Key:      "linux",
		Strategy: intent.Strategy{Primary: "df -h {{path}}"},
	}
	got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{Shell: "bash"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Argv) < 3 || got.Argv[2] != "." {
		t.Fatalf("argv %v", got.Argv)
	}
}

func TestRenderRejectsControlChars(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{
		Intent: "search_text_in_file",
		Params: map[string]string{"pattern": "bad\nline", "file": "logs.txt"},
	}
	selected := capabilities.SelectedStrategy{
		Strategy: intent.Strategy{Primary: "grep {{pattern}} {{file}}"},
	}
	_, err := Render(context.Background(), resolved, selected, environment.SystemProfile{})
	if err == nil {
		t.Fatal("expected error")
	}
}
