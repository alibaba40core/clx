package aliases

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreSetListRemove(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "aliases.yaml")
	ctx := context.Background()

	s, err := OpenAt(ctx, path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set(ctx, "gst", "git status"); err != nil {
		t.Fatal(err)
	}
	v, ok := s.Lookup("GST")
	if !ok || v != "git status" {
		t.Fatalf("lookup=%q ok=%v", v, ok)
	}
	if err := s.Remove(ctx, "gst"); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Lookup("gst"); ok {
		t.Fatal("expected removed")
	}
}

func TestStoreRejectInvalidName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := OpenAt(context.Background(), filepath.Join(dir, "aliases.yaml"), 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set(context.Background(), "bad name", "x"); err != ErrInvalidName {
		t.Fatalf("err=%v", err)
	}
}

func TestReadMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.yaml")
	s, err := OpenAt(context.Background(), path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 0 {
		t.Fatal("expected empty")
	}
}

func TestEncodeRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "aliases.yaml")
	ctx := context.Background()
	s, err := OpenAt(ctx, path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set(ctx, "prd", "cd /tmp"); err != nil {
		t.Fatal(err)
	}
	s2, err := OpenAt(ctx, path, 10)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := s2.Lookup("prd")
	if !ok || v != "cd /tmp" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "prd") {
		t.Fatalf("file=%s", data)
	}
}
