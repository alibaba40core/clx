package parser

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func testProfileWindows() environment.SystemProfile {
	return environment.SystemProfile{
		OS:    "windows",
		Shell: "powershell",
	}
}

func testProfileLinux() environment.SystemProfile {
	return environment.SystemProfile{
		OS:    "linux",
		Shell: "bash",
	}
}

func TestParseMatrix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name     string
		input    string
		profile  environment.SystemProfile
		wantType InputType
		wantTok  []string
		wantArgs map[string]string
	}{
		{
			name:     "partial grep windows",
			input:    `grep errors logs.txt`,
			profile:  testProfileWindows(),
			wantType: InputPartialShell,
			wantTok:  []string{"grep", "errors", "logs.txt"},
		},
		{
			name:     "shell grep linux",
			input:    `grep errors logs.txt`,
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"grep", "errors", "logs.txt"},
		},
		{
			name:     "partial locate windows",
			input:    "locate help.txt",
			profile:  testProfileWindows(),
			wantType: InputPartialShell,
			wantTok:  []string{"locate", "help.txt"},
		},
		{
			name:     "shell locate linux",
			input:    "locate help.txt",
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"locate", "help.txt"},
		},
		{
			name:     "natural language",
			input:    "find all files modified today",
			profile:  testProfileWindows(),
			wantType: InputNaturalLanguage,
			wantTok:  []string{"find", "all", "files", "modified", "today"},
		},
		{
			name:     "clx invocation",
			input:    "clx grep errors logs.txt",
			profile:  testProfileLinux(),
			wantType: InputCLXInvocation,
			wantTok:  []string{"grep", "errors", "logs.txt"},
		},
		{
			name:     "clxmax invocation",
			input:    "clxmax grep errors logs.txt",
			profile:  testProfileLinux(),
			wantType: InputCLXInvocation,
			wantTok:  []string{"grep", "errors", "logs.txt"},
		},
		{
			name:     "clx doctor",
			input:    "clx doctor",
			profile:  testProfileWindows(),
			wantType: InputCLXInvocation,
			wantTok:  []string{"doctor"},
		},
		{
			name:     "git status",
			input:    "git status",
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"git", "status"},
		},
		{
			name:     "env assignment",
			input:    "FOO=bar grep x",
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"grep", "x"},
			wantArgs: map[string]string{"FOO": "bar"},
		},
		{
			name:     "clx nl body still invocation",
			input:    "clx find all files modified today",
			profile:  testProfileWindows(),
			wantType: InputCLXInvocation,
			wantTok:  []string{"find", "all", "files", "modified", "today"},
		},
		{
			name:     "ping linux flags not nl",
			input:    "ping -c 4 google.com",
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"ping", "-c", "4", "google.com"},
		},
		{
			name:     "ping windows flags not nl",
			input:    "ping -n 4 google.com",
			profile:  testProfileWindows(),
			wantType: InputShell,
			wantTok:  []string{"ping", "-n", "4", "google.com"},
		},
		{
			name:     "netstat listening not nl",
			input:    "netstat -tlnp",
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"netstat", "-tlnp"},
		},
		{
			name:     "ss listening not nl",
			input:    "ss -tlnp",
			profile:  testProfileLinux(),
			wantType: InputShell,
			wantTok:  []string{"ss", "-tlnp"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(ctx, tc.input, tc.profile, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got.InputType != tc.wantType {
				t.Fatalf("type: got %s want %s", got.InputType, tc.wantType)
			}
			if len(got.Tokens) != len(tc.wantTok) {
				t.Fatalf("tokens: got %v want %v", got.Tokens, tc.wantTok)
			}
			for i := range tc.wantTok {
				if got.Tokens[i] != tc.wantTok[i] {
					t.Fatalf("token[%d]: got %q want %q", i, got.Tokens[i], tc.wantTok[i])
				}
			}
			if tc.wantArgs != nil {
				for k, v := range tc.wantArgs {
					if got.Args[k] != v {
						t.Fatalf("arg %s: got %q want %q", k, got.Args[k], v)
					}
				}
			}
		})
	}
}

func TestParseEmptyInput(t *testing.T) {
	t.Parallel()
	_, err := Parse(context.Background(), "  ", testProfileLinux(), nil)
	if err != errEmptyInput {
		t.Fatalf("got %v", err)
	}
}

func TestParseContextCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Parse(ctx, "grep x", testProfileLinux(), nil)
	if err != context.Canceled {
		t.Fatalf("got %v", err)
	}
}
