package config

import (
	"path/filepath"
	"testing"
)

func TestHomeCLXHomeOverride(t *testing.T) {
	t.Setenv("CLX_HOME", filepath.Clean("/tmp/clx-test-home"))
	home, err := Home()
	if err != nil {
		t.Fatal(err)
	}
	if home != "/tmp/clx-test-home" && home != `C:\tmp\clx-test-home` {
		// Windows may clean differently; accept filepath.Clean result.
		want := filepath.Clean("/tmp/clx-test-home")
		if home != want {
			t.Fatalf("Home() = %q, want %q", home, want)
		}
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	p, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "config.yaml" {
		t.Fatalf("got %s", p)
	}
}
