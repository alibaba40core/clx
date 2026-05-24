package executor

import (
	"testing"
)

func TestCleanAndValidatePathRelative(t *testing.T) {
	t.Parallel()
	got, err := CleanAndValidatePath("logs.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "logs.txt" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanAndValidatePathRejectsTraversal(t *testing.T) {
	t.Parallel()
	_, err := CleanAndValidatePath("../etc/passwd")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCleanAndValidatePathRejectsMetachar(t *testing.T) {
	t.Parallel()
	_, err := CleanAndValidatePath("file;rm")
	if err == nil {
		t.Fatal("expected error")
	}
}
