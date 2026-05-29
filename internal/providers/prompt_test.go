package providers

import (
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func sampleRequest() IntentRequest {
	return IntentRequest{
		RawInput: "find all .log files",
		Profile: environment.SystemProfile{
			OS:       "windows",
			Shell:    "powershell",
			Terminal: "Windows Terminal",
			AvailableTools: []string{
				"git", "vim", "node", "rg",
			},
		},
		KnownIntents: []string{"list_dir", "find_file", "current_dir"},
		RuleParams: map[string][]string{
			"find_file":   {"filename", "path"},
			"list_dir":    {"path"},
			"current_dir": {},
		},
	}
}

func TestBuildPromptIncludesProfile(t *testing.T) {
	t.Parallel()
	_, user, err := BuildPrompt(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"OS: windows", "Shell: powershell", "Terminal: Windows Terminal"} {
		if !strings.Contains(user, want) {
			t.Fatalf("user message missing %q:\n%s", want, user)
		}
	}
	if !strings.Contains(user, `Path hint: use "." for the current directory, not "/".`) {
		t.Fatalf("missing Windows path hint:\n%s", user)
	}
}

func TestBuildPromptSortsIntents(t *testing.T) {
	t.Parallel()
	req := sampleRequest()
	req.KnownIntents = []string{"zebra", "alpha", "middle"}
	_, user, err := BuildPrompt(req)
	if err != nil {
		t.Fatal(err)
	}
	idxAlpha := strings.Index(user, "Allowed intents:")
	if idxAlpha < 0 {
		t.Fatal("missing Allowed intents section")
	}
	section := user[idxAlpha:]
	if !strings.Contains(section, "alpha, middle, zebra") {
		t.Fatalf("intents not sorted: %s", section)
	}
}

func TestBuildPromptIncludesIntentParameters(t *testing.T) {
	t.Parallel()
	_, user, err := BuildPrompt(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(user, "Intent parameters") {
		t.Fatalf("missing intent parameters section: %s", user)
	}
	if !strings.Contains(user, "- find_file: filename, path") {
		t.Fatalf("missing find_file params: %s", user)
	}
	if !strings.Contains(user, "- current_dir: none") {
		t.Fatalf("missing current_dir params: %s", user)
	}
}

func TestBuildPromptFiltersTools(t *testing.T) {
	t.Parallel()
	_, user, err := BuildPrompt(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(user, "git") || !strings.Contains(user, "rg") {
		t.Fatalf("expected allowlisted tools in user message: %s", user)
	}
	if strings.Contains(user, "vim") || strings.Contains(user, "node") {
		t.Fatalf("non-allowlisted tools must not appear: %s", user)
	}
}

func TestBuildPromptRedactsRawInput(t *testing.T) {
	t.Parallel()
	req := sampleRequest()
	req.RawInput = "deploy with api_key=sk-abc123 now"
	_, user, err := BuildPrompt(req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(user, "sk-abc123") {
		t.Fatalf("secret leaked into prompt: %s", user)
	}
	if !strings.Contains(user, "[REDACTED]") {
		t.Fatalf("expected redaction marker in prompt: %s", user)
	}
}

func TestBuildPromptBoundedSize(t *testing.T) {
	t.Parallel()
	req := sampleRequest()
	req.RawInput = strings.Repeat("x", maxPromptUserBytes)
	_, _, err := BuildPrompt(req)
	if err == nil {
		t.Fatal("expected error for oversized prompt")
	}
	if !strings.Contains(err.Error(), "prompt too large") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPromptRejectsTooManyIntents(t *testing.T) {
	t.Parallel()
	req := sampleRequest()
	req.KnownIntents = make([]string, maxKnownIntents+1)
	for i := range req.KnownIntents {
		req.KnownIntents[i] = "intent"
	}
	_, _, err := BuildPrompt(req)
	if err == nil {
		t.Fatal("expected error for too many intents")
	}
	if !strings.Contains(err.Error(), "too many known intents") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPromptDeterministic(t *testing.T) {
	t.Parallel()
	req := sampleRequest()
	req.KnownIntents = []string{"b", "a", "c"}
	s1, u1, err := BuildPrompt(req)
	if err != nil {
		t.Fatal(err)
	}
	s2, u2, err := BuildPrompt(req)
	if err != nil {
		t.Fatal(err)
	}
	if s1 != s2 || u1 != u2 {
		t.Fatalf("non-deterministic output:\nfirst:  %q\nsecond: %q", u1, u2)
	}
}

func TestBuildPromptSystemMessageConstant(t *testing.T) {
	t.Parallel()
	system, _, err := BuildPrompt(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if system != systemPrompt {
		t.Fatalf("system = %q, want constant", system)
	}
}
