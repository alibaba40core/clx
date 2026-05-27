package environment

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestRunDoctorWritesProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)

	var out bytes.Buffer
	if err := RunDoctor(context.Background(), &out, DoctorOptions{}); err != nil {
		t.Fatal(err)
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("profile missing: %v", err)
	}

	store, err := LoadStore(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if store.SchemaVersion != SchemaVersion {
		t.Fatalf("schema %d", store.SchemaVersion)
	}
	key := ProfileKey(detectOS(), detectShell())
	profile, ok := store.Profiles[key]
	if !ok {
		t.Fatalf("missing profile for key %q", key)
	}
	if profile.OS == "" || profile.Shell == "" {
		t.Fatalf("incomplete profile: %+v", profile)
	}
	if profile.Paths["home"] == "" {
		t.Fatal("expected home path")
	}
	if !bytes.Contains(out.Bytes(), []byte("environment profile written")) {
		t.Fatalf("output: %s", out.String())
	}
}

func TestRunDoctorRefreshPreservesSiblingShells(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	ctx := context.Background()

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}

	curShell := detectShell()
	siblingShell := "cmd"
	if curShell == "cmd" {
		siblingShell = "powershell"
	}

	store := NewProfileStore()
	store.UpsertProfile(SystemProfile{
		OS:             detectOS(),
		Shell:          siblingShell,
		OSVersion:      "sibling-preserved",
		AvailableTools: []string{"legacy"},
	})
	if err := SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}

	if err := RunDoctor(ctx, ioDiscard{}, DoctorOptions{Refresh: true}); err != nil {
		t.Fatal(err)
	}

	store2, err := LoadStore(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	siblingKey := ProfileKey(detectOS(), siblingShell)
	sibling, ok := store2.Profiles[siblingKey]
	if !ok {
		t.Fatalf("sibling %q profile removed", siblingKey)
	}
	if sibling.OSVersion != "sibling-preserved" {
		t.Fatalf("sibling mutated: %+v", sibling)
	}

	currentKey := ProfileKey(detectOS(), curShell)
	current, ok := store2.Profiles[currentKey]
	if !ok {
		t.Fatal("current shell profile missing")
	}
	if current.OSVersion == "sibling-preserved" {
		t.Fatal("current shell should have been re-detected")
	}
	if current.Shell != curShell {
		t.Fatalf("shell %q", current.Shell)
	}
}

func TestDetectHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Detect(ctx)
	if err != context.Canceled {
		t.Fatalf("got %v", err)
	}
}

func TestDoctorProfilePathUnderCLXHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	_ = RunDoctor(context.Background(), ioDiscard{}, DoctorOptions{})
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != filepath.Clean(home) {
		t.Fatalf("profile %s not under %s", path, home)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
