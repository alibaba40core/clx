package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}

func TestClxHookScriptsMentionExplain(t *testing.T) {
	root := repoRoot(t)
	for _, name := range []string{"clx-hook.sh", "clx-hook.ps1"} {
		data, err := os.ReadFile(filepath.Join(root, "scripts", name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "--explain") {
			t.Fatalf("%s should use explain-only mode", name)
		}
	}
}

func TestClxHookShSyntax(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash -n on windows")
	}
	sh := filepath.Join(repoRoot(t), "scripts", "clx-hook.sh")
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not installed")
	}
	cmd := exec.Command("bash", "-n", sh)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bash -n: %v %s", err, out)
	}
}
