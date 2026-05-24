package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

func TestCheckBlocked(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLX_HOME", dir)
	ResetCache()

	polDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(polDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(polDir, "policy.yaml"), []byte("blocked:\n  - \"rm -rf\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "/"}, Command: "rm -rf /"}
	got, err := Check(context.Background(), gen, risk.RiskAssessment{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Allowed {
		t.Fatal("expected blocked")
	}
}
