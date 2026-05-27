package executor

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildValidatedScript(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		shell   string
		argv    []string
		wantErr error
		wantSub string
	}{
		{
			name:    "powershell cmdlet",
			shell:   "powershell",
			argv:    []string{"Get-Location"},
			wantSub: "Get-Location",
		},
		{
			name:    "powershell multi",
			shell:   "powershell",
			argv:    []string{"Select-String", "errors", "logs.txt"},
			wantSub: "Select-String",
		},
		{
			name:    "posix safe",
			shell:   "bash",
			argv:    []string{"grep", "pattern", "file.txt"},
			wantSub: "grep",
		},
		{
			name:    "cmd",
			shell:   "cmd",
			argv:    []string{"cd"},
			wantSub: "cd",
		},
		{
			name:    "empty argv",
			shell:   "powershell",
			argv:    nil,
			wantErr: ErrEmptyScriptArgv,
		},
		{
			name:    "semicolon injection",
			shell:   "powershell",
			argv:    []string{"Get-Location; rm -rf /"},
			wantErr: ErrScriptMetachar,
		},
		{
			name:    "pipe injection",
			shell:   "bash",
			argv:    []string{"grep", "a|b"},
			wantErr: ErrScriptMetachar,
		},
		{
			name:    "dollar injection",
			shell:   "powershell",
			argv:    []string{"Write-Host", "$(whoami)"},
			wantErr: ErrScriptMetachar,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := BuildValidatedScript(tc.shell, tc.argv)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatal("expected error")
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err %v want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantSub != "" && !strings.Contains(got, tc.wantSub) {
				t.Fatalf("script %q missing %q", got, tc.wantSub)
			}
		})
	}
}
