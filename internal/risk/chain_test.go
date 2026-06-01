package risk

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
)

func TestAssessChainExprMediumFloor(t *testing.T) {
	gen := generator.GeneratedCommand{
		Chain: &generator.CommandChain{
			Stages: []generator.ChainStage{
				{Tokens: []generator.ChainToken{{Value: "Get-ChildItem"}}},
				{Tokens: []generator.ChainToken{
					{Value: "Where-Object"},
					{Value: "{ $_.x -eq 1 }", Expr: true},
				}},
			},
			Connectors: []generator.ChainConnector{generator.ChainPipe},
		},
	}
	ra, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if ra.Level != Medium {
		t.Fatalf("expected medium, got %s", ra.Level)
	}
}

func TestAssessChainDestructiveTailHigh(t *testing.T) {
	gen := generator.GeneratedCommand{
		Chain: &generator.CommandChain{
			Stages: []generator.ChainStage{
				{Tokens: []generator.ChainToken{{Value: "Get-ChildItem"}}},
				{Tokens: []generator.ChainToken{{Value: "rm"}, {Value: "-rf"}, {Value: "."}}},
			},
			Connectors: []generator.ChainConnector{generator.ChainPipe},
		},
	}
	ra, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if ra.Level != High {
		t.Fatalf("expected high, got %s (%s)", ra.Level, ra.Reason)
	}
}
