package executor

import (
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func TestValidateCommandQualityRejectsPlaceholders(t *testing.T) {
	t.Parallel()
	cases := []string{"URL", "file", "linkName", "file.tar.gz", "."}
	for _, tok := range cases {
		gen := generator.NewAICommand([]string{"curl", "-o", tok}, "powershell", "x", testProfile())
		if err := ValidateCommandQuality(gen, "download"); err == nil {
			t.Fatalf("token %q should be rejected", tok)
		} else if !strings.Contains(err.Error(), "placeholder") {
			t.Fatalf("unexpected error for %q: %v", tok, err)
		}
	}
}

func TestValidateCommandQualityAcceptsValidWhereObjectChain(t *testing.T) {
	t.Parallel()
	chain := &generator.CommandChain{
		Stages: []generator.ChainStage{
			{Tokens: []generator.ChainToken{
				{Value: "Get-NetTCPConnection"},
				{Value: "-LocalPort"},
				{Value: "443"},
			}},
			{Tokens: []generator.ChainToken{
				{Value: "Where-Object", Expr: false},
				{Value: "{$_.State -eq 'Listen'}", Expr: true},
			}},
		},
		Connectors: []generator.ChainConnector{generator.ChainPipe},
	}
	gen := generator.GeneratedCommand{
		Shell: "powershell",
		Chain: chain,
	}
	if err := ValidateCommandQuality(gen, "port 443 listening"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCommandQualityRejectsBareScriptblockStage(t *testing.T) {
	t.Parallel()
	chain := &generator.CommandChain{
		Stages: []generator.ChainStage{
			{Tokens: []generator.ChainToken{{Value: "Get-NetTCPConnection"}}},
			{Tokens: []generator.ChainToken{
				{Value: "{ $_.State -eq 'Listen' }", Expr: true},
			}},
		},
		Connectors: []generator.ChainConnector{generator.ChainPipe},
	}
	gen := generator.GeneratedCommand{Shell: "powershell", Chain: chain}
	if err := ValidateCommandQuality(gen, "port 443"); err == nil {
		t.Fatal("expected broken filter rejection")
	}
}

func TestValidateCommandQualityCompareFolders(t *testing.T) {
	t.Parallel()
	gen := generator.NewAICommand([]string{"Get-ChildItem", "C:\\a"}, "powershell", "x", testProfile())
	if err := ValidateCommandQuality(gen, "compare two folders"); err == nil {
		t.Fatal("expected compare quality rejection")
	}
	chain := &generator.CommandChain{
		Stages: []generator.ChainStage{
			{Tokens: []generator.ChainToken{
				{Value: "Compare-Object"},
				{Value: "(Get-ChildItem C:\\a)"},
				{Value: "(Get-ChildItem C:\\b)"},
			}},
		},
	}
	gen2 := generator.GeneratedCommand{Shell: "powershell", Chain: chain}
	if err := ValidateCommandQuality(gen2, "compare two folders"); err != nil {
		t.Fatal(err)
	}
}

func testProfile() environment.SystemProfile {
	return environment.SystemProfile{Shell: "powershell"}
}
