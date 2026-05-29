package providers

import (
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func TestBuildExplainPromptDeterministic(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{
		Argv:  []string{"git", "status"},
		Shell: "bash",
	}
	profile := environment.SystemProfile{OS: "linux", Shell: "bash"}
	sys1, user1, err := BuildExplainPrompt(gen, profile)
	if err != nil {
		t.Fatal(err)
	}
	sys2, user2, err := BuildExplainPrompt(gen, profile)
	if err != nil {
		t.Fatal(err)
	}
	if sys1 != sys2 || user1 != user2 {
		t.Fatal("expected deterministic prompt")
	}
	if !strings.Contains(user1, "git") {
		t.Fatalf("user=%q", user1)
	}
}

func TestBuildExplainPromptRedactsSecrets(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{
		Argv:  []string{"curl", "-H", "Authorization: Bearer sk-secret1234567890"},
		Shell: "bash",
	}
	profile := environment.SystemProfile{OS: "linux", Shell: "bash"}
	_, user, err := BuildExplainPrompt(gen, profile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(user, "sk-secret1234567890") {
		t.Fatalf("secret leaked in prompt: %q", user)
	}
}
