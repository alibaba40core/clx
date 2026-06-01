package executor

import (
	"errors"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func sampleChain() *generator.CommandChain {
	return &generator.CommandChain{
		Stages: []generator.ChainStage{
			{Tokens: []generator.ChainToken{{Value: "Get-ChildItem"}, {Value: "-Recurse"}}},
			{Tokens: []generator.ChainToken{
				{Value: "Where-Object"},
				{Value: "{ $_.LastWriteTime -ge (Get-Date).Date }", Expr: true},
			}},
		},
		Connectors: []generator.ChainConnector{generator.ChainPipe},
	}
}

func TestValidateCommandChainAccepts(t *testing.T) {
	if err := ValidateCommandChain(sampleChain(), "powershell"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCommandChainRejectsSingleStage(t *testing.T) {
	c := &generator.CommandChain{Stages: []generator.ChainStage{
		{Tokens: []generator.ChainToken{{Value: "ls"}}},
	}}
	if err := ValidateCommandChain(c, "bash"); !errors.Is(err, ErrChainEmpty) {
		t.Fatalf("got %v", err)
	}
}

func TestBuildValidatedChainScript(t *testing.T) {
	prof := environment.SystemProfile{Shell: "powershell", ShellVersion: "7.4"}
	script, err := BuildValidatedChainScript("powershell", sampleChain(), prof)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "|") {
		t.Fatalf("expected pipe: %q", script)
	}
	if !strings.Contains(script, "$_.LastWriteTime") {
		t.Fatalf("expr verbatim: %q", script)
	}
}

func TestBuildValidatedChainScriptAndConnector(t *testing.T) {
	c := &generator.CommandChain{
		Stages: []generator.ChainStage{
			{Tokens: []generator.ChainToken{{Value: "echo"}, {Value: "hello"}}},
			{Tokens: []generator.ChainToken{{Value: "echo"}, {Value: "world"}}},
		},
		Connectors: []generator.ChainConnector{generator.ChainAnd},
	}
	prof := environment.SystemProfile{Shell: "bash"}
	script, err := BuildValidatedChainScript("bash", c, prof)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, "&&") {
		t.Fatalf("expected &&: %q", script)
	}
}
