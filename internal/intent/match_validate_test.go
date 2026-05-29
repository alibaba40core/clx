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

func TestValidateResolvedParamsRejectsMissingRequiredFromExamples(t *testing.T) {
	rule := Rule{
		Intent:   "find_file",
		Examples: []string{"locate {{filename}}", "find file {{filename}} in {{path}}"},
	}
	err := validateResolvedParams(rule, map[string]string{})
	if err == nil {
		t.Fatal("expected missing filename")
	}
}

func TestValidateResolvedParamsAllowsOptionalPathOnListDir(t *testing.T) {
	rule := Rule{
		Intent:   "list_dir",
		Examples: []string{"ls {{path}}", "ll"},
	}
	if err := validateResolvedParams(rule, map[string]string{}); err != nil {
		t.Fatal(err)
	}
}
