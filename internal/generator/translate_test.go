package generator

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

func testEngine(t *testing.T) *intent.Engine {
	t.Helper()
	eng, err := intent.NewEngineFromModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

func argvEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestTranslateSeedIntents(t *testing.T) {
	t.Parallel()
	eng := testEngine(t)
	ctx := context.Background()

	cases := []struct {
		name    string
		resolved intent.ResolvedIntent
		profile environment.SystemProfile
		wantArgv []string
	}{
		{
			name: "search linux rg",
			resolved: intent.ResolvedIntent{
				Intent: "search_text_in_file",
				Params: map[string]string{"pattern": "errors", "file": "logs.txt"},
			},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash", AvailableTools: []string{"rg"}},
			wantArgv: []string{"rg", "errors", "logs.txt"},
		},
		{
			name: "search linux grep",
			resolved: intent.ResolvedIntent{
				Intent: "search_text_in_file",
				Params: map[string]string{"pattern": "errors", "file": "logs.txt"},
			},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash", AvailableTools: []string{"grep"}},
			wantArgv: []string{"grep", "errors", "logs.txt"},
		},
		{
			name: "search powershell",
			resolved: intent.ResolvedIntent{
				Intent: "search_text_in_file",
				Params: map[string]string{"pattern": "errors", "file": "logs.txt"},
			},
			profile:  environment.SystemProfile{OS: "windows", Shell: "powershell"},
			wantArgv: []string{"Select-String", "errors", "logs.txt"},
		},
		{
			name: "find_file powershell",
			resolved: intent.ResolvedIntent{
				Intent: "find_file",
				Params: map[string]string{"filename": "help.txt"},
			},
			profile: environment.SystemProfile{OS: "windows", Shell: "powershell"},
			wantArgv: []string{"Get-ChildItem", "-Recurse", "-Filter", "help.txt"},
		},
		{
			name: "find_file linux find",
			resolved: intent.ResolvedIntent{
				Intent: "find_file",
				Params: map[string]string{"filename": "help.txt"},
			},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash"},
			wantArgv: []string{"find", ".", "-name", "help.txt"},
		},
		{
			name: "find_file linux fd",
			resolved: intent.ResolvedIntent{
				Intent: "find_file",
				Params: map[string]string{"filename": "help.txt"},
			},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash", AvailableTools: []string{"fd"}},
			wantArgv: []string{"fd", ".", "-t", "f", "-g", "help.txt"},
		},
		{
			name: "list_dir powershell",
			resolved: intent.ResolvedIntent{
				Intent: "list_dir",
				Params: map[string]string{"path": "C:\\tmp"},
			},
			profile:  environment.SystemProfile{OS: "windows", Shell: "powershell"},
			wantArgv: []string{"Get-ChildItem", "C:\\tmp"},
		},
		{
			name: "list_dir linux",
			resolved: intent.ResolvedIntent{
				Intent: "list_dir",
				Params: map[string]string{"path": "/tmp"},
			},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash"},
			wantArgv: []string{"ls", "-la", "/tmp"},
		},
		{
			name: "current_dir linux",
			resolved: intent.ResolvedIntent{Intent: "current_dir", Params: map[string]string{}},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash"},
			wantArgv: []string{"pwd"},
		},
		{
			name: "current_dir powershell",
			resolved: intent.ResolvedIntent{Intent: "current_dir", Params: map[string]string{}},
			profile:  environment.SystemProfile{OS: "windows", Shell: "powershell"},
			wantArgv: []string{"Get-Location"},
		},
		{
			name: "disk_usage linux default path",
			resolved: intent.ResolvedIntent{Intent: "disk_usage", Params: map[string]string{}},
			profile:  environment.SystemProfile{OS: "linux", Shell: "bash"},
			wantArgv: []string{"df", "-h", "."},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Translate(ctx, eng, tc.resolved, tc.profile)
			if err != nil {
				t.Fatal(err)
			}
			if !argvEqual(got.Argv, tc.wantArgv) {
				t.Fatalf("argv %v want %v", got.Argv, tc.wantArgv)
			}
			if got.Shell != tc.profile.Shell {
				t.Fatalf("shell %q", got.Shell)
			}
			if got.Explanation == "" {
				t.Fatal("missing explanation")
			}
		})
	}
}
