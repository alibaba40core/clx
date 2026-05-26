package environment

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrDetectCreatesProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)

	p, err := LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.OS == "" {
		t.Fatal("expected OS in profile")
	}

	path := filepath.Join(dir, "system_profile.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("profile not written: %v", err)
	}

	p2, err := LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p2.OS != p.OS || p2.Shell != p.Shell {
		t.Fatalf("reload got %+v want %+v", p2, p)
	}
}

func TestLoadOrDetectCacheHitPerShell(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ctx := context.Background()

	p1, err := LoadOrDetect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Mutate on-disk store; cached load must still return original without re-read corruption.
	path := filepath.Join(dir, "system_profile.json")
	store, err := LoadStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	key := ProfileKey(p1.OS, p1.Shell)
	stale := store.Profiles[key]
	stale.OSVersion = "stale-version"
	store.Profiles[key] = stale
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}

	p2, err := LoadOrDetect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p2.OSVersion != "stale-version" {
		t.Fatalf("expected cache hit with on-disk value, got %q", p2.OSVersion)
	}
}

func TestLoadOrDetectMissWhenShellChanges(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ctx := context.Background()

	path := filepath.Join(dir, "system_profile.json")
	store := NewProfileStore()
	store.UpsertProfile(SystemProfile{
		OS:             detectOS(),
		Shell:          "cmd",
		OSVersion:      "cached-cmd",
		AvailableTools: []string{"git"},
	})
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}

	// Simulate PowerShell session.
	t.Setenv("SHELL", "")
	t.Setenv("MSYSTEM", "")
	t.Setenv("ComSpec", `C:\Windows\System32\cmd.exe`)
	t.Setenv("PSModulePath", "C:\\Modules")
	t.Setenv("POWERSHELL_VERSION", "7.4.0")

	if detectShell() != "powershell" {
		t.Skip("environment is not PowerShell after env setup")
	}

	p, err := LoadOrDetect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if p.Shell != "powershell" {
		t.Fatalf("got shell %q", p.Shell)
	}
	if p.OSVersion == "cached-cmd" {
		t.Fatal("expected fresh detect for powershell, got cmd cache")
	}

	store2, err := LoadStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store2.Profiles[ProfileKey(detectOS(), "cmd")]; !ok {
		t.Fatal("cmd profile should remain after powershell detect")
	}
}

func TestLoadOrDetectMigratesV1(t *testing.T) {
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

	p, err := LoadOrDetect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if p.OSVersion != "v1-migrated" {
		t.Fatalf("got OSVersion %q", p.OSVersion)
	}

	store, err := LoadStore(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if store.SchemaVersion != SchemaVersion {
		t.Fatalf("schema %d", store.SchemaVersion)
	}
	if len(store.Profiles) == 0 {
		t.Fatal("expected profiles map")
	}
}
