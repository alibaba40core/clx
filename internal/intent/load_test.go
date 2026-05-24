package intent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRulesFromModuleRoot(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngineFromModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if len(eng.rules) < 5 {
		t.Fatalf("expected at least 5 rules, got %d", len(eng.rules))
	}
	_ = root
}

func TestLoadRulesFromFS(t *testing.T) {
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	rules, err := LoadRulesFromFS(os.DirFS(root), "rules")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range rules {
		if r.Intent == "find_file" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("find_file not loaded")
	}
	_ = filepath.Join(root, "rules")
}
