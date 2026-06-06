package environment

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfileMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)

	_, err := LoadProfile(context.Background())
	if err != ErrProfileNotFound {
		t.Fatalf("got %v want ErrProfileNotFound", err)
	}
}

func TestLoadProfileReadsPersisted(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ctx := context.Background()

	path := filepath.Join(dir, "system_profile.json")
	store := NewProfileStore()
	store.UpsertProfile(SystemProfile{
		OS:        detectOS(),
		Shell:     detectShell(),
		OSVersion: "saved",
	})
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}

	p, err := LoadProfile(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p.OSVersion != "saved" {
		t.Fatalf("got %q", p.OSVersion)
	}
}

func TestLoadOrDetectDelegatesToLoadProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)

	_, err := LoadOrDetect(context.Background())
	if err != ErrProfileNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestLoadProfileMigratesV1(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	path := filepath.Join(dir, "system_profile.json")

	v1 := SystemProfile{
		SchemaVersion:  1,
		OS:             detectOS(),
		Shell:          detectShell(),
		OSVersion:      "v1-migrated",
		AvailableTools: []string{"git"},
	}
	data, err := json.Marshal(v1)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	p, err := LoadProfile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.OSVersion != "v1-migrated" {
		t.Fatalf("got OSVersion %q", p.OSVersion)
	}
}

func TestMinimalProfileHasOSAndShell(t *testing.T) {
	p := MinimalProfile()
	if p.OS == "" || p.Shell == "" {
		t.Fatalf("got %+v", p)
	}
}
