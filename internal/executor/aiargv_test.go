package executor

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateGeneratedArgvAccepts(t *testing.T) {
	cases := [][]string{
		{"ls", "-la"},
		{"dir", "."},
		{"del", "*"}, // glob is allowed; risk engine flags it as destructive
		{"git", "status"},
		{"grep", "-r", "needle", "."},
	}
	for _, argv := range cases {
		if err := ValidateGeneratedArgv(argv, "bash"); err != nil {
			t.Errorf("argv %v: unexpected error %v", argv, err)
		}
	}
}

func TestValidateGeneratedArgvRejectsMetachars(t *testing.T) {
	cases := [][]string{
		{"ls", ";", "rm"},
		{"ls", "|", "sh"},
		{"ls", "&&", "rm"},
		{"echo", "$(whoami)"},
		{"echo", "`whoami`"},
		{"cat", "x", ">", "y"},
		{"cat", "<", "y"},
		{"foo;bar"},
	}
	for _, argv := range cases {
		err := ValidateGeneratedArgv(argv, "bash")
		if !errors.Is(err, ErrAIArgvToken) {
			t.Errorf("argv %v: expected ErrAIArgvToken, got %v", argv, err)
		}
	}
}

func TestValidateGeneratedArgvCmdRejectsPercent(t *testing.T) {
	if err := ValidateGeneratedArgv([]string{"echo", "%PATH%"}, "cmd"); !errors.Is(err, ErrAIArgvToken) {
		t.Fatalf("expected ErrAIArgvToken for %%VAR%% on cmd, got %v", err)
	}
	// On bash, % is harmless and allowed.
	if err := ValidateGeneratedArgv([]string{"echo", "100%"}, "bash"); err != nil {
		t.Fatalf("bash should allow %%, got %v", err)
	}
}

func TestValidateGeneratedArgvEmpty(t *testing.T) {
	if err := ValidateGeneratedArgv(nil, "bash"); !errors.Is(err, ErrAIArgvEmpty) {
		t.Fatalf("expected ErrAIArgvEmpty, got %v", err)
	}
	if err := ValidateGeneratedArgv([]string{"ls", ""}, "bash"); !errors.Is(err, ErrAIArgvToken) {
		t.Fatalf("expected ErrAIArgvToken for empty token, got %v", err)
	}
}

func TestValidateGeneratedArgvTooLong(t *testing.T) {
	argv := make([]string, maxAIArgvTokens+1)
	for i := range argv {
		argv[i] = "x"
	}
	if err := ValidateGeneratedArgv(argv, "bash"); !errors.Is(err, ErrAIArgvTooLong) {
		t.Fatalf("expected ErrAIArgvTooLong, got %v", err)
	}
}

func TestValidateGeneratedArgvNullByte(t *testing.T) {
	if err := ValidateGeneratedArgv([]string{"ls", "a\x00b"}, "bash"); !errors.Is(err, ErrAIArgvToken) {
		t.Fatalf("expected ErrAIArgvToken for null byte, got %v", err)
	}
}

func TestValidateGeneratedArgvTokenTooLong(t *testing.T) {
	long := strings.Repeat("a", maxAIArgvTokenBytes+1)
	if err := ValidateGeneratedArgv([]string{"echo", long}, "bash"); !errors.Is(err, ErrAIArgvToken) {
		t.Fatalf("expected ErrAIArgvToken for oversized token, got %v", err)
	}
}
