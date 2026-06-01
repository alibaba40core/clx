package shellchain

import (
	"testing"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

func TestSymbolPipe(t *testing.T) {
	s, err := Symbol(generator.ChainPipe, environment.SystemProfile{Shell: "bash"})
	if err != nil || s != "|" {
		t.Fatalf("got %q err=%v", s, err)
	}
}

func TestSymbolAndBash(t *testing.T) {
	s, err := Symbol(generator.ChainAnd, environment.SystemProfile{Shell: "bash"})
	if err != nil || s != "&&" {
		t.Fatalf("got %q err=%v", s, err)
	}
}

func TestSymbolAndCmd(t *testing.T) {
	s, err := Symbol(generator.ChainAnd, environment.SystemProfile{Shell: "cmd"})
	if err != nil || s != "&" {
		t.Fatalf("got %q err=%v", s, err)
	}
}

func TestSymbolAndPowerShell5(t *testing.T) {
	s, err := Symbol(generator.ChainAnd, environment.SystemProfile{Shell: "powershell", ShellVersion: "5.1"})
	if err != nil || s != ";" {
		t.Fatalf("got %q err=%v", s, err)
	}
}

func TestSymbolAndPowerShell7(t *testing.T) {
	s, err := Symbol(generator.ChainAnd, environment.SystemProfile{Shell: "pwsh", ShellVersion: "7.4.0"})
	if err != nil || s != "&&" {
		t.Fatalf("got %q err=%v", s, err)
	}
}
