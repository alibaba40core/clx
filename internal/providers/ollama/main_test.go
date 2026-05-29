package ollama

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain enforces C10 from doc/phase-2.md: any package that spawns goroutines
// (HTTP client conn pools, httptest servers) must verify no goroutine leaks at
// test teardown. A leaked Keep-Alive idle conn or unclosed httptest goroutine
// would silently regress runtime-footprint.mdc's "Goroutines at idle: <=3" budget
// without showing up in any other gate.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
