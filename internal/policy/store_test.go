package policy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAddRemoveAllowedVerb(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	ctx := context.Background()
	if err := os.MkdirAll(filepath.Join(dir, "policies"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := AddAllowedVerb(ctx, "pwd"); err != nil {
		t.Fatal(err)
	}
	list, err := ListAllowedVerbs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0] != "pwd" {
		t.Fatalf("list=%v", list)
	}
	if err := RemoveAllowedVerb(ctx, "pwd"); err != nil {
		t.Fatal(err)
	}
	list, err = ListAllowedVerbs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("list=%v", list)
	}
}

func TestEnsureHighDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	ctx := context.Background()
	if err := EnsureHighDefaults(ctx); err != nil {
		t.Fatal(err)
	}
	list, err := ListAllowedVerbs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultHighAllowedVerbs()
	if len(list) != len(want) {
		t.Fatalf("list=%v want %v", list, want)
	}
}

func TestAddAllowedVerbInvalid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	if err := AddAllowedVerb(context.Background(), "rm -rf"); !errors.Is(err, ErrVerbInvalid) {
		t.Fatalf("got %v", err)
	}
}

func TestSetAccessLevelAndBlockedPatterns(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	ctx := context.Background()
	if err := os.MkdirAll(filepath.Join(dir, "policies"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := SetAccessLevel(ctx, "moderate"); err != nil {
		t.Fatal(err)
	}
	pol, err := Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pol.AccessLevel != AccessModerate {
		t.Fatalf("access=%v", pol.AccessLevel)
	}
	if err := AddBlockedPattern(ctx, "rm -rf"); err != nil {
		t.Fatal(err)
	}
	list, err := ListBlockedPatterns(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0] != "rm -rf" {
		t.Fatalf("list=%v", list)
	}
	if err := RemoveBlockedPattern(ctx, "rm -rf"); err != nil {
		t.Fatal(err)
	}
	list, err = ListBlockedPatterns(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("list=%v", list)
	}
}

func TestAddBlockedPatternInvalid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	if err := AddBlockedPattern(context.Background(), "bad|pipe"); !errors.Is(err, ErrPatternInvalid) {
		t.Fatalf("got %v", err)
	}
}
