package risk

import (
	"context"
	"testing"

	"github.com/alibaba40core/clx/internal/generator"
)

func TestAssessLowSeed(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"grep", "x", "y"}, Command: "grep x y"}
	got, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != Low {
		t.Fatalf("level %v", got.Level)
	}
}

func TestAssessHighDestructive(t *testing.T) {
	t.Parallel()
	gen := generator.GeneratedCommand{Argv: []string{"rm", "-rf", "/"}, Command: "rm -rf /"}
	got, err := Assess(context.Background(), gen)
	if err != nil {
		t.Fatal(err)
	}
	if got.Level != High {
		t.Fatalf("level %v", got.Level)
	}
}
