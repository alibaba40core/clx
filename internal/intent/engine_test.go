package intent

import (
	"sort"
	"testing"
)

func TestNewDefaultEngineMatchesModuleRoot(t *testing.T) {
	embedded, err := NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	fromFS, err := NewEngineFromModuleRoot()
	if err != nil {
		t.Skipf("module root not available: %v", err)
	}

	gotEmbedded := intentNames(embedded)
	gotFS := intentNames(fromFS)
	sort.Strings(gotEmbedded)
	sort.Strings(gotFS)

	if len(gotEmbedded) != len(gotFS) {
		t.Fatalf("intent count: embedded %d, module root %d", len(gotEmbedded), len(gotFS))
	}
	for i := range gotEmbedded {
		if gotEmbedded[i] != gotFS[i] {
			t.Fatalf("intent mismatch at %d: embedded %q, module root %q", i, gotEmbedded[i], gotFS[i])
		}
	}
}

func intentNames(eng *Engine) []string {
	names := make([]string, 0, len(eng.rules))
	for _, r := range eng.rules {
		names = append(names, r.Intent)
	}
	return names
}
