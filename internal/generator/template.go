package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
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

// normalizePathForProfile maps AI/resolver path mistakes to values safe on the target OS.
// On Windows, "/" is not a valid relative path for cmd's dir and is treated as a switch.
func normalizePathForProfile(path string, profile environment.SystemProfile) string {
	if strings.ToLower(profile.OS) != "windows" {
		return path
	}
	switch path {
	case "/", "//":
		return "."
	}
	return path
}

// effectiveParams merges resolved params with defaults for optional keys.
// Defaults exist so that a bare invocation (e.g. `git log` with no -n) can
// still render a single template containing the parameter placeholder.
func effectiveParams(intentName string, params map[string]string, profile environment.SystemProfile) map[string]string {
	out := make(map[string]string, len(params)+1)
	for k, v := range params {
		out[k] = v
	}
	if path, ok := out["path"]; ok {
		out["path"] = normalizePathForProfile(path, profile)
	}
	if _, ok := out["path"]; !ok {
		switch intentName {
		case "disk_usage", "list_dir", "find_file", "find_modified_today", "find_large_files":
			out["path"] = "."
		}
	}
	if _, ok := out["size"]; !ok {
		if intentName == "find_large_files" {
			out["size"] = "100M"
		}
	}
	if intentName == "find_large_files" {
		sizeBytes, findSize, err := normalizeFileSizeParams(out["size"])
		if err == nil {
			out["size"] = findSize
			out["size_bytes"] = strconv.FormatInt(sizeBytes, 10)
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

// normalizeFileSizeParams converts user size tokens (e.g. 100MB) into find(1) suffix
// form and byte count for PowerShell length comparisons.
func normalizeFileSizeParams(size string) (bytes int64, findSize string, err error) {
	s := strings.TrimSpace(strings.ToUpper(size))
	if s == "" {
		return 0, "", fmt.Errorf("empty size")
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "GB"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		mult = 1024
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "G"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "G")
	case strings.HasSuffix(s, "M"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "K"):
		mult = 1024
		s = strings.TrimSuffix(s, "K")
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || n <= 0 {
		return 0, "", fmt.Errorf("invalid size %q", size)
	}
	bytes = n * mult
	findSize = fmt.Sprintf("%d%c", n, sizeSuffixForFind(mult))
	return bytes, findSize, nil
}

func sizeSuffixForFind(mult int64) byte {
	switch mult {
	case 1024:
		return 'K'
	case 1024 * 1024:
		return 'M'
	case 1024 * 1024 * 1024:
		return 'G'
	default:
		return 'M'
	}
}
