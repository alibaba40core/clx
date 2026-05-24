package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFS(t *testing.T) {
	engRoot := findRoot(t)
	rules, err := LoadFromFS(os.DirFS(engRoot), "skills")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rules {
		if r.Intent == "list_dir" {
			return
		}
	}
	t.Fatal("list_dir not found in skills")
}

func findRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
