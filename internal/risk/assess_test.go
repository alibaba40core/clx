package risk

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
)

func TestAssessLowSeed(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"grep", "x", "y"}, Command: "grep x y"}
	got, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != Low {
		t.Fatalf("level %v", got.Level)
	}
}

func TestAssessHighDestructive(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "/"}, Command: "rm -rf /"}
	got, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != High {
		t.Fatalf("level %v", got.Level)
	}
}

func TestAssessLowVerbsMatrix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		argv []string
	}{
		{"docker_ps", []string{"docker", "ps"}},
		{"docker_images", []string{"docker", "images"}},
		{"docker_logs", []string{"docker", "logs", "--tail", "200", "web"}},
		{"git_status", []string{"git", "status"}},
		{"git_log", []string{"git", "log", "--oneline", "-n", "20"}},
		{"git_diff", []string{"git", "diff"}},
		{"git_branch", []string{"git", "branch"}},
		{"curl_head", []string{"curl", "-I", "https://example.com"}},
		{"wget_dryrun", []string{"wget", "--spider", "https://example.com"}},
		{"ping_linux", []string{"ping", "-c", "4", "google.com"}},
		{"ping_windows", []string{"ping", "-n", "4", "google.com"}},
		{"ss_listening", []string{"ss", "-tlnp"}},
		{"netstat_listening", []string{"netstat", "-an"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gen := generator.GeneratedCommand{Argv: tc.argv, Command: joinArgv(tc.argv)}
			got, err := Assess(context.Background(), gen)
			if err != nil {
				t.Fatal(err)
			}
			if got.Level != Low {
				t.Fatalf("%s: level %v reason=%q", tc.name, got.Level, got.Reason)
			}
			if got.RequiresConfirmation {
				t.Fatalf("%s: should not require confirmation", tc.name)
			}
		})
	}
}

func TestAssessGitNonReadOnlySubverbMedium(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"git", "push"},
		{"git", "commit", "-m", "x"},
		{"git", "reset", "--hard"},
		{"git", "checkout", "main"},
	}
	for _, argv := range cases {
		gen := generator.GeneratedCommand{Argv: argv, Command: joinArgv(argv)}
		got, err := Assess(context.Background(), gen)
		if err != nil {
			t.Fatal(err)
		}
		if got.Level != Medium {
			t.Fatalf("%v: level %v", argv, got.Level)
		}
		if !got.RequiresConfirmation {
			t.Fatalf("%v: should require confirmation", argv)
		}
	}
}

func TestAssessDockerNonReadOnlySubverbMedium(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"docker", "run", "alpine"},
		{"docker", "build", "."},
		{"docker", "exec", "-it", "web", "sh"},
		{"docker", "pull", "alpine"},
		{"docker", "tag", "alpine:latest", "alpine:dev"},
	}
	for _, argv := range cases {
		gen := generator.GeneratedCommand{Argv: argv, Command: joinArgv(argv)}
		got, err := Assess(context.Background(), gen)
		if err != nil {
			t.Fatal(err)
		}
		if got.Level != Medium {
			t.Fatalf("%v: level %v reason=%q", argv, got.Level, got.Reason)
		}
	}
}

func TestAssessDockerRmStillHigh(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{
		Argv:    []string{"docker", "rm", "-f", "web"},
		Command: "docker rm -f web",
	}
	got, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != High {
		t.Fatalf("expected high for docker rm, got %v", got.Level)
	}
}

func joinArgv(a []string) string {
	out := ""
	for i, s := range a {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}
