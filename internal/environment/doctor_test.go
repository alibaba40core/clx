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
	if err := RunDoctor(context.Background(), &out); err != nil {
		t.Fatal(err)
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("profile missing: %v", err)
	}

	profile, err := Load(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if profile.SchemaVersion != SchemaVersion {
		t.Fatalf("schema %d", profile.SchemaVersion)
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
	_ = RunDoctor(context.Background(), ioDiscard{})
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
