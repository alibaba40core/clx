package intent

import (
	"errors"
	"testing"
)

func TestMissErrorIsErrNotFound(t *testing.T) {
	t.Parallel()
	err := &MissError{AIAttempted: true}
	if !errors.Is(err, ErrNotFound) {
		t.Fatal("MissError should match ErrNotFound")
	}
}

func TestAsMiss(t *testing.T) {
	t.Parallel()
	m, ok := AsMiss(&MissError{AIAttempted: true})
	if !ok || !m.AIAttempted {
		t.Fatalf("AsMiss = %v %v", m, ok)
	}
}
