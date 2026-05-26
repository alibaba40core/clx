package environment

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoadStoreRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "system_profile.json")

	orig := SystemProfile{
		OS:             "linux",
		OSVersion:      "22.04",
		Shell:          "bash",
		ShellVersion:   "5.1",
		Terminal:       "unknown",
		PackageManagers: []string{"apt"},
		AvailableTools: []string{"git", "go"},
		Paths:          map[string]string{"home": "/home/user", "workspace": "/tmp"},
	}

	store := NewProfileStore()
	store.UpsertProfile(orig)
	store.UpsertProfile(SystemProfile{
		OS:    "windows",
		Shell: "cmd",
	})

	ctx := context.Background()
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}
	got, err := LoadStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Fatalf("schema %d want %d", got.SchemaVersion, SchemaVersion)
	}
	p := got.Profiles[ProfileKey("linux", "bash")]
	if p.OS != orig.OS || p.Shell != orig.Shell || len(p.AvailableTools) != 2 {
		t.Fatalf("linux profile %+v", p)
	}
	if _, ok := got.Profiles[ProfileKey("windows", "cmd")]; !ok {
		t.Fatal("missing windows-cmd profile")
	}
}

func TestSaveStoreOverwritesEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "system_profile.json")
	ctx := context.Background()

	store := NewProfileStore()
	store.UpsertProfile(SystemProfile{OS: "linux", Shell: "bash"})
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}

	store.UpsertProfile(SystemProfile{OS: "linux", Shell: "bash", OSVersion: "24.04"})
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}
	got, err := LoadStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	p := got.Profiles[ProfileKey("linux", "bash")]
	if p.OSVersion != "24.04" {
		t.Fatalf("expected overwrite, got %+v", p)
	}
}

func TestLoadStoreMissingFile(t *testing.T) {
	t.Parallel()
	_, err := LoadStore(context.Background(), filepath.Join(t.TempDir(), "missing.json"))
	if !os.IsNotExist(err) {
		t.Fatalf("expected not exist, got %v", err)
	}
}

func TestMigrateV1FlatProfile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "system_profile.json")

	v1 := `{
  "schema_version": 1,
  "os": "windows",
  "os_version": "10.0",
  "shell": "powershell",
  "terminal": "unknown",
  "package_managers": ["winget"],
  "available_tools": ["git"],
  "wsl_enabled": true,
  "paths": {"home": "C:\\Users\\u"}
}`
	if err := os.WriteFile(path, []byte(v1), 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := LoadStore(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if store.SchemaVersion != SchemaVersion {
		t.Fatalf("schema %d", store.SchemaVersion)
	}
	p, ok := store.Profiles[ProfileKey("windows", "powershell")]
	if !ok {
		t.Fatal("missing migrated profile")
	}
	if p.OS != "windows" || p.Shell != "powershell" {
		t.Fatalf("got %+v", p)
	}
	if len(p.AvailableTools) != 1 || p.AvailableTools[0] != "git" {
		t.Fatalf("tools %+v", p.AvailableTools)
	}

	// Migration should have been persisted.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"profiles"`) || !strings.Contains(body, `"windows-powershell"`) {
		t.Fatalf("file not migrated: %s", raw)
	}
}

func TestProfileKey(t *testing.T) {
	t.Parallel()
	if got := ProfileKey("Windows", "PowerShell"); got != "windows-powershell" {
		t.Fatalf("got %q", got)
	}
}
