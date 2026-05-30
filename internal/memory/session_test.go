package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestAppendBounded(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ctx := context.Background()
	cfg := config.MemoryConfig{MaxEntriesPerSession: 2}
	s, err := Open(ctx, "test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if err := s.AppendCommand(ctx, CommandEntry{RawInput: "x", Intent: "list_dir"}); err != nil {
			t.Fatal(err)
		}
	}
	s.mu.Lock()
	n := len(s.sess.Commands)
	s.mu.Unlock()
	if n != 2 {
		t.Fatalf("len=%d", n)
	}
}

func TestCorruptFileFailsOpen(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	if _, err := config.Bootstrap(context.Background()); err != nil {
		t.Fatal(err)
	}
	path, err := SessionPath("bad")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = Open(context.Background(), "bad", config.Default().Memory)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestDefaultSessionID(t *testing.T) {
	t.Setenv("CLX_SESSION_ID", "")
	if DefaultSessionID() != "default" {
		t.Fatal()
	}
	t.Setenv("CLX_SESSION_ID", "my-session")
	if DefaultSessionID() != "my-session" {
		t.Fatal()
	}
}

func TestSessionPath(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	p, err := SessionPath("abc")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "abc.json" {
		t.Fatalf("%s", p)
	}
}
