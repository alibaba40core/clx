package capabilities

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

func searchRule() intent.Rule {
	return intent.Rule{
		Intent: "search_text_in_file",
		Strategies: map[string]intent.Strategy{
			"linux":      {Primary: "grep {{pattern}} {{file}}"},
			"linux_rg":   {Primary: "rg {{pattern}} {{file}}", RequiresTool: "rg", Priority: 10},
			"powershell": {Argv: []string{"Select-String", "{{pattern}}", "{{file}}"}},
			"cmd":        {Primary: "findstr {{pattern}} {{file}}"},
		},
	}
}

func TestSelectLinuxWithRg(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "linux", Shell: "bash", AvailableTools: []string{"rg", "grep"}}
	got, err := Select(context.Background(), searchRule(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "linux_rg" {
		t.Fatalf("key %q", got.Key)
	}
}

func TestSelectLinuxWithoutRg(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "linux", Shell: "bash", AvailableTools: []string{"grep"}}
	got, err := Select(context.Background(), searchRule(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "linux" {
		t.Fatalf("key %q", got.Key)
	}
}

func TestSelectPowerShell(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "windows", Shell: "powershell"}
	got, err := Select(context.Background(), searchRule(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "powershell" {
		t.Fatalf("key %q", got.Key)
	}
}

func TestSelectCmd(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "windows", Shell: "cmd"}
	got, err := Select(context.Background(), searchRule(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "cmd" {
		t.Fatalf("key %q", got.Key)
	}
}

func TestSelectNoStrategy(t *testing.T) {
	t.Parallel()
	rule := intent.Rule{Intent: "empty", Strategies: map[string]intent.Strategy{}}
	profile := environment.SystemProfile{OS: "linux", Shell: "bash"}
	_, err := Select(context.Background(), rule, profile)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStrategyMatchesKeyLinuxPrefix(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "linux", Shell: "bash"}
	if !strategyMatchesKey("linux_rg", profile) {
		t.Fatal("linux_rg should match linux profile")
	}
	if strategyMatchesKey("powershell", profile) {
		t.Fatal("powershell should not match linux profile")
	}
}
