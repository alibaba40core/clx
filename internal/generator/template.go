package generator

import (
	"fmt"
	"strings"
)

func substituteSlot(slot string, params map[string]string) (string, error) {
	if !strings.Contains(slot, "{{") {
		return slot, nil
	}
	out := slot
	for k, v := range params {
		if err := validateParamValue(v); err != nil {
			return "", err
		}
		placeholder := "{{" + k + "}}"
		out = strings.ReplaceAll(out, placeholder, v)
	}
	if strings.Contains(out, "{{") {
		return "", fmt.Errorf("unsubstituted placeholder in %q", slot)
	}
	return out, nil
}

func validateParamValue(v string) error {
	if v == "" {
		return nil
	}
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c < 0x20 && c != '\t' {
			return fmt.Errorf("invalid character in parameter value")
		}
	}
	return nil
}

// effectiveParams merges resolved params with defaults for optional keys.
// Defaults exist so that a bare invocation (e.g. `git log` with no -n) can
// still render a single template containing the parameter placeholder.
func effectiveParams(intentName string, params map[string]string) map[string]string {
	out := make(map[string]string, len(params)+1)
	for k, v := range params {
		out[k] = v
	}
	if _, ok := out["path"]; !ok {
		switch intentName {
		case "disk_usage", "list_dir":
			out["path"] = "."
		}
	}
	if _, ok := out["n"]; !ok {
		if intentName == "git_log" {
			out["n"] = "20"
		}
	}
	if _, ok := out["lines"]; !ok {
		if intentName == "docker_logs" {
			out["lines"] = "200"
		}
	}
	return out
}
