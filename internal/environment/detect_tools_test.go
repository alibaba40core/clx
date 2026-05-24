package environment

import (
	"testing"
)

func TestFindOnPathMock(t *testing.T) {
	old := lookPath
	defer func() { lookPath = old }()

	lookPath = func(name string) (string, error) {
		if name == "git" || name == "go" {
			return "/usr/bin/" + name, nil
		}
		return "", errNotFound{}
	}

	found := findOnPath([]string{"git", "docker", "go"})
	if len(found) != 2 || found[0] != "git" || found[1] != "go" {
		t.Fatalf("got %v", found)
	}
}

type errNotFound struct{}

func (errNotFound) Error() string { return "not found" }

func TestToolCatalogBounded(t *testing.T) {
	if len(toolCatalog) > 32 {
		t.Fatalf("tool catalog exceeds bound: %d", len(toolCatalog))
	}
}
