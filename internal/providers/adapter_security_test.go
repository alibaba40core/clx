package providers

import (
	"context"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

// Security tests: adapter output must pass Engine.ValidateResolved when wired through pipeline.
// Here we verify malicious provider output is structurally rejected before exec would run.

func TestAdapterRejectsUnknownIntentViaPipelineValidation(t *testing.T) {
	t.Parallel()
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	r := AsResolver(stubProvider{resp: &IntentResponse{
		Intent: "rm_rf_slash", Params: map[string]string{}, Confidence: 0.99,
	}}, eng, nil, AdapterConfig{})
	resolved, err := r.Resolve(context.Background(), parser.Request{RawInput: "delete everything"})
	if err != nil {
		t.Fatalf("adapter should return AI hit: %v", err)
	}
	valErr := eng.ValidateResolved(resolved)
	if valErr == nil {
		t.Fatal("ValidateResolved must reject unknown intent rm_rf_slash")
	}
	if !strings.Contains(valErr.Error(), "unknown intent") {
		t.Fatalf("err = %v", valErr)
	}
}

func TestAdapterRejectsExtraParamViaPipelineValidation(t *testing.T) {
	t.Parallel()
	eng, err := intent.NewDefaultEngine()
	if err != nil {
		t.Fatal(err)
	}
	// list_dir has no declared params in builtin rules
	r := AsResolver(stubProvider{resp: &IntentResponse{
		Intent: "list_dir", Params: map[string]string{"evil": "x"}, Confidence: 0.99,
	}}, eng, nil, AdapterConfig{})
	resolved, err := r.Resolve(context.Background(), parser.Request{RawInput: "ls"})
	if err != nil {
		t.Fatalf("adapter resolve: %v", err)
	}
	if err := eng.ValidateResolved(resolved); err == nil {
		t.Fatal("ValidateResolved must reject extra param")
	}
}
