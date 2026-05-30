package providers

import (
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

func TestBuildPromptIncludesSkillHints(t *testing.T) {
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	hints := SkillPromptsForEngine(eng)
	if len(hints) == 0 {
		t.Fatal("expected filesystem skill prompt")
	}
	_, user, err := BuildPrompt(IntentRequest{
		RawInput:     "list files here",
		Profile:      environment.SystemProfile{OS: "linux", Shell: "bash"},
		KnownIntents: []string{"list_dir"},
		SkillHints:   hints,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(user, "Domain hints:") || !strings.Contains(user, "filesystem") {
		t.Fatalf("user prompt missing skill hints: %q", user)
	}
}
