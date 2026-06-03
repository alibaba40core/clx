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

func TestRenderFindFileDefaultPath(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{
		Intent: "find_file",
		Params: map[string]string{"filename": "clx.exe"},
	}
	cases := []struct {
		key      string
		strategy intent.Strategy
		want     []string
	}{
		{
			key:      "linux",
			strategy: intent.Strategy{Primary: "find {{path}} -name {{filename}}"},
			want:     []string{"find", ".", "-name", "clx.exe"},
		},
		{
			key:      "linux_fd",
			strategy: intent.Strategy{Primary: "fd -t f -g {{filename}} {{path}}"},
			want:     []string{"fd", "-t", "f", "-g", "clx.exe", "."},
		},
		{
			key: "powershell",
			strategy: intent.Strategy{
				Argv: []string{
					"Get-ChildItem", "-Path", "{{path}}", "-Recurse", "-Filter", "{{filename}}",
					"-ErrorAction", "SilentlyContinue",
				},
			},
			want: []string{
				"Get-ChildItem", "-Path", ".", "-Recurse", "-Filter", "clx.exe",
				"-ErrorAction", "SilentlyContinue",
			},
		},
		{
			key:      "cmd",
			strategy: intent.Strategy{Primary: "dir /s {{path}}\\{{filename}}"},
			want:     []string{"dir", "/s", ".\\clx.exe"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			selected := capabilities.SelectedStrategy{Key: tc.key, Strategy: tc.strategy}
			got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{Shell: tc.key})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Argv) != len(tc.want) {
				t.Fatalf("argv %v want %v", got.Argv, tc.want)
			}
			for i := range tc.want {
				if got.Argv[i] != tc.want[i] {
					t.Fatalf("argv[%d]=%q want %q (full %v)", i, got.Argv[i], tc.want[i], got.Argv)
				}
			}
		})
	}
}

func TestRenderFindFileExplicitPathPowerShell(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{
		Intent: "find_file",
		Params: map[string]string{"filename": "clx.exe", "path": `C:\foo`},
	}
	selected := capabilities.SelectedStrategy{
		Key: "powershell",
		Strategy: intent.Strategy{
			Argv: []string{
				"Get-ChildItem", "-Path", "{{path}}", "-Recurse", "-Filter", "{{filename}}",
				"-ErrorAction", "SilentlyContinue",
			},
		},
	}
	got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{Shell: "powershell"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Get-ChildItem", "-Path", `C:\foo`, "-Recurse", "-Filter", "clx.exe",
		"-ErrorAction", "SilentlyContinue",
	}
	for i := range want {
		if got.Argv[i] != want[i] {
			t.Fatalf("argv[%d]=%q want %q (full %v)", i, got.Argv[i], want[i], got.Argv)
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

func TestRenderFindModifiedTodayChain(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{Intent: "find_modified_today", Params: map[string]string{}}
	selected := capabilities.SelectedStrategy{
		Key: "powershell",
		Strategy: intent.Strategy{
			Chain: &intent.ChainSpec{
				Stages: []intent.ChainStageSpec{
					{Argv: []string{"Get-ChildItem", "-Recurse", "-File"}},
					{Tokens: []intent.ChainTokenSpec{
						{Value: "Where-Object"},
						{Value: "{ $_.LastWriteTime.Date -eq [datetime]::Today }", Expr: true},
					}},
				},
				Connectors: []string{"pipe"},
			},
		},
	}
	got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{Shell: "powershell"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Chain == nil || len(got.Chain.Stages) != 2 {
		t.Fatalf("expected chain, got %+v", got.Chain)
	}
	if got.ExecHost != ExecPowerShell {
		t.Fatalf("exec host=%v", got.ExecHost)
	}
}

func TestRenderListEnvPowerShell(t *testing.T) {
	t.Parallel()
	resolved := intent.ResolvedIntent{Intent: "list_env", Params: map[string]string{}}
	selected := capabilities.SelectedStrategy{
		Key:      "powershell",
		Strategy: intent.Strategy{Primary: "Get-ChildItem env:"},
	}
	got, err := Render(context.Background(), resolved, selected, environment.SystemProfile{Shell: "powershell"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Get-ChildItem", "env:"}
	if len(got.Argv) != len(want) {
		t.Fatalf("argv %v", got.Argv)
	}
	for i := range want {
		if got.Argv[i] != want[i] {
			t.Fatalf("argv[%d]=%q want %q", i, got.Argv[i], want[i])
		}
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
