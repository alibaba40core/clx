package generator

import (
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func TestNormalizePathForProfileWindowsRoot(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "windows", Shell: "cmd"}
	for _, in := range []string{"/", "//"} {
		if got := normalizePathForProfile(in, profile); got != "." {
			t.Fatalf("normalizePathForProfile(%q) = %q, want .", in, got)
		}
	}
}

func TestNormalizePathForProfilePreservesLinuxRoot(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "linux", Shell: "bash"}
	if got := normalizePathForProfile("/", profile); got != "/" {
		t.Fatalf("got %q want /", got)
	}
}

func TestEffectiveParamsNormalizesWindowsRootPath(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "windows", Shell: "cmd"}
	got := effectiveParams("list_dir", map[string]string{"path": "/"}, profile)
	if got["path"] != "." {
		t.Fatalf("path = %q, want .", got["path"])
	}
}

func TestEffectiveParamsFindLargeFilesDefaults(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "windows", Shell: "powershell"}
	got := effectiveParams("find_large_files", map[string]string{}, profile)
	if got["path"] != "." || got["size"] != "100M" || got["size_bytes"] != "104857600" {
		t.Fatalf("got %v", got)
	}
}

func TestEffectiveParamsFindLargeFilesParsesMB(t *testing.T) {
	t.Parallel()
	profile := environment.SystemProfile{OS: "linux", Shell: "bash"}
	got := effectiveParams("find_large_files", map[string]string{"size": "100MB"}, profile)
	if got["size"] != "100M" || got["size_bytes"] != "104857600" {
		t.Fatalf("got %v", got)
	}
}
