package providers

import (
	"errors"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func sampleProfile() environment.SystemProfile {
	return environment.SystemProfile{
		OS:              "windows",
		OSVersion:       "10.0.22631",
		Shell:           "cmd",
		Terminal:        "unknown",
		WSLEnabled:      true,
		PackageManagers: []string{"npm", "winget"},
		AvailableTools:  []string{"git", "docker", "go", "curl"},
	}
}

func TestBuildCommandPromptIncludesPlatform(t *testing.T) {
	system, user, err := BuildCommandPrompt(CommandRequest{
		RawInput: "remove all files in this dir",
		Profile:  sampleProfile(),
	})
	if err != nil {
		t.Fatalf("BuildCommandPrompt: %v", err)
	}
	if system != commandSystemPrompt {
		t.Fatalf("system prompt mismatch: %q", system)
	}
	for _, want := range []string{
		"OS: windows 10.0.22631",
		"Shell: cmd",
		"WSL: available",
		"Package managers: npm, winget",
		"Available tools:",
		"git",
		"Request: remove all files in this dir",
		"Path hint:",
	} {
		if !strings.Contains(user, want) {
			t.Errorf("user message missing %q\n---\n%s", want, user)
		}
	}
}

func TestBuildCommandPromptRedactsSecrets(t *testing.T) {
	_, user, err := BuildCommandPrompt(CommandRequest{
		RawInput: "use token sk-ABCDEFGHIJKLMNOPQRSTUVWXYZ012345",
		Profile:  sampleProfile(),
	})
	if err != nil {
		t.Fatalf("BuildCommandPrompt: %v", err)
	}
	if strings.Contains(user, "sk-ABCDEFGHIJKLMNOPQRSTUVWXYZ012345") {
		t.Fatalf("secret leaked into prompt: %s", user)
	}
}

func TestBuildCommandSchemaShape(t *testing.T) {
	schema := BuildCommandSchema()
	if schema["additionalProperties"] != false {
		t.Fatalf("expected additionalProperties false")
	}
	req, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("required not a slice: %T", schema["required"])
	}
	want := map[string]bool{"argv": false, "shell": false, "explanation": false, "confidence": false}
	for _, r := range req {
		if s, ok := r.(string); ok {
			want[s] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("required missing %q", k)
		}
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties missing")
	}
	argv, ok := props["argv"].(map[string]any)
	if !ok || argv["type"] != "array" {
		t.Fatalf("argv not an array schema: %v", props["argv"])
	}
}

func TestParseCommandContentValid(t *testing.T) {
	resp, err := ParseCommandContent(`{"argv":["dir","."],"shell":"cmd","explanation":"list","confidence":0.9}`)
	if err != nil {
		t.Fatalf("ParseCommandContent: %v", err)
	}
	if len(resp.Argv) != 2 || resp.Argv[0] != "dir" || resp.Argv[1] != "." {
		t.Fatalf("argv=%v", resp.Argv)
	}
	if resp.Shell != "cmd" || resp.Confidence != 0.9 {
		t.Fatalf("resp=%+v", resp)
	}
}

func TestParseCommandContentFiltersEmptyTokens(t *testing.T) {
	resp, err := ParseCommandContent(`{"argv":["ls","","-la"],"shell":"bash"}`)
	if err != nil {
		t.Fatalf("ParseCommandContent: %v", err)
	}
	if len(resp.Argv) != 2 || resp.Argv[1] != "-la" {
		t.Fatalf("argv=%v", resp.Argv)
	}
}

func TestParseCommandContentNoArgv(t *testing.T) {
	if _, err := ParseCommandContent(`{"argv":[],"shell":"bash"}`); !errors.Is(err, ErrNoMatch) {
		t.Fatalf("expected ErrNoMatch, got %v", err)
	}
}

func TestParseCommandContentBadJSON(t *testing.T) {
	if _, err := ParseCommandContent(`not json`); !errors.Is(err, ErrInvalidResp) {
		t.Fatalf("expected ErrInvalidResp, got %v", err)
	}
}
