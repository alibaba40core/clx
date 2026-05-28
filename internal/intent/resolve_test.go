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

func TestResolveFindFileWithPath(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	req := parser.Request{Tokens: []string{"locate", "clx.exe", "in", `C:\foo`}}
	got, err := eng.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if got.Intent != "find_file" {
		t.Fatalf("intent: got %q want find_file", got.Intent)
	}
	if got.Params["filename"] != "clx.exe" {
		t.Fatalf("filename: got %q want clx.exe", got.Params["filename"])
	}
	if got.Params["path"] != `C:\foo` {
		t.Fatalf("path: got %q want C:\\foo", got.Params["path"])
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

func TestResolveDailyUsageIntents(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	cases := []struct {
		name       string
		tokens     []string
		wantIntent string
		wantParams map[string]string
	}{
		{"ll_bare", []string{"ll"}, "list_dir", map[string]string{}},
		{"ll_path", []string{"ll", "/tmp"}, "list_dir", map[string]string{"path": "/tmp"}},
		{"ip_bare", []string{"ip"}, "show_ip_addresses", map[string]string{}},
		{"ipconfig", []string{"ipconfig"}, "show_ip_addresses", map[string]string{}},
		{"rm_file", []string{"rm", "test"}, "remove_file", map[string]string{"path": "test"}},
		{"rmdir", []string{"rmdir", "text"}, "remove_dir", map[string]string{"path": "text"}},
		{"remove_file", []string{"remove", "test"}, "remove_file", map[string]string{"path": "test"}},
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
			for k, v := range tc.wantParams {
				if got.Params[k] != v {
					t.Fatalf("param %q: got %q want %q", k, got.Params[k], v)
				}
			}
		})
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

func TestResolveNetworkingIntents(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	cases := []struct {
		name       string
		tokens     []string
		wantIntent string
		wantParams map[string]string
	}{
		{"ping_host", []string{"ping", "google.com"}, "ping_host", map[string]string{"host": "google.com"}},
		{"curl_url", []string{"curl", "-I", "https://example.com"}, "curl_url", map[string]string{"url": "https://example.com"}},
		{"netstat_listening_ss", []string{"ss", "-tlnp"}, "netstat_listening", map[string]string{}},
		{"netstat_listening_netstat", []string{"netstat", "-tlnp"}, "netstat_listening", map[string]string{}},
		{"netstat_listening_netstat_an", []string{"netstat", "-an"}, "netstat_listening", map[string]string{}},
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

func TestResolveDockerIntents(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	cases := []struct {
		name       string
		tokens     []string
		wantIntent string
		wantParams map[string]string
	}{
		{"docker_ps", []string{"docker", "ps"}, "docker_ps", map[string]string{}},
		{"docker_images", []string{"docker", "images"}, "docker_images", map[string]string{}},
		{"docker_logs_bare", []string{"docker", "logs", "web"}, "docker_logs", map[string]string{"container": "web"}},
		{"docker_logs_tail", []string{"docker", "logs", "--tail", "100", "web"}, "docker_logs", map[string]string{"container": "web", "lines": "100"}},
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
