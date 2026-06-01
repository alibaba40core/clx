package parser

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
)

func TestParseChainedShell(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	profile := environment.SystemProfile{OS: "windows", Shell: "powershell"}
	raw := `Get-ChildItem -Recurse | Where-Object { $_.Name -eq "x" }`
	req, err := Parse(ctx, raw, profile, nil)
	if err != nil {
		t.Fatal(err)
	}
	if req.InputType != InputChainedShell {
		t.Fatalf("type=%v", req.InputType)
	}
	if req.ShellChain == nil || len(req.ShellChain.Stages) != 2 {
		t.Fatalf("chain=%+v", req.ShellChain)
	}
}
