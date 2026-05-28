package intent

import "testing"

func TestValidateResolvedParamsRejectsExtraFromExamples(t *testing.T) {
	rule := Rule{
		Intent:   "find_file",
		Examples: []string{"locate {{filename}}", "locate {{filename}} in {{path}}"},
	}
	err := validateResolvedParams(rule, map[string]string{
		"filename":   "x",
		"evil_extra": "y",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateResolvedParamsAcceptsKnownKeys(t *testing.T) {
	rule := Rule{
		Intent:   "find_file",
		Examples: []string{"locate {{filename}}"},
	}
	err := validateResolvedParams(rule, map[string]string{"filename": "x"})
	if err != nil {
		t.Fatal(err)
	}
}
