package environment

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "system_profile.json")

	orig := SystemProfile{
		SchemaVersion:   SchemaVersion,
		OS:              "linux",
		OSVersion:       "22.04",
		Shell:           "bash",
		ShellVersion:    "5.1",
		Terminal:        "unknown",
		PackageManagers: []string{"apt"},
		AvailableTools:  []string{"git", "go"},
		WSLEnabled:      false,
		Paths:           map[string]string{"home": "/home/user", "workspace": "/tmp"},
	}

	ctx := context.Background()
	if err := Save(ctx, path, orig); err != nil {
		t.Fatal(err)
	}
	got, err := Load(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if got.OS != orig.OS || got.Shell != orig.Shell || len(got.AvailableTools) != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestSaveOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "system_profile.json")
	ctx := context.Background()

	if err := Save(ctx, path, SystemProfile{OS: "linux"}); err != nil {
		t.Fatal(err)
	}
	if err := Save(ctx, path, SystemProfile{OS: "windows"}); err != nil {
		t.Fatal(err)
	}
	got, err := Load(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if got.OS != "windows" {
		t.Fatalf("expected overwrite, got %q", got.OS)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()
	_, err := Load(context.Background(), filepath.Join(t.TempDir(), "missing.json"))
	if !os.IsNotExist(err) {
		t.Fatalf("expected not exist, got %v", err)
	}
}
