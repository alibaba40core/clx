package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/risk"
)

func TestCheckChainBlocksDestructiveStage(t *testing.T) {
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

	gen := generator.GeneratedCommand{
		Chain: &generator.CommandChain{
			Stages: []generator.ChainStage{
				{Tokens: []generator.ChainToken{{Value: "ls"}}},
				{Tokens: []generator.ChainToken{{Value: "rm"}, {Value: "-rf"}, {Value: "/"}}},
			},
			Connectors: []generator.ChainConnector{generator.ChainPipe},
		},
	}
	got, err := Check(context.Background(), gen, risk.RiskAssessment{Level: risk.High}, CheckOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecAllowed {
		t.Fatal("expected block on chain stage 2")
	}
}
