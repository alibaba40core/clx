package test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/aliases"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/pipeline"
	"github.com/alibaba40core/clx/internal/providers"
)

type aliasStubAI struct{}

func (aliasStubAI) Resolve(context.Context, parser.Request) (intent.ResolvedIntent, error) {
	return intent.ResolvedIntent{}, intent.ErrNotFound
}

func TestE2EAliasDestructiveBlockedByPolicy(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	writePolicy(t, []string{"rm -rf"})

	ctx := context.Background()
	store, err := aliases.Open(ctx, 32)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(ctx, "boom", "rm -rf /"); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Safety.Mode = "low"
	cfg.Execution.AutoExecute = true

	var stderr bytes.Buffer
	code, err := pipeline.Run(ctx, cfg, "boom", pipeline.Options{
		Yes:        true,
		AliasStore: store,
		AIResolver: aliasStubAI{},
		Provider: &fakeCmdProvider{resp: &providers.CommandResponse{
			Argv:        []string{"rm", "-rf", "/"},
			Shell:       "bash",
			Explanation: "destructive",
			Confidence:  0.9,
		}},
		Stderr: &stderr,
		Stdout: &bytes.Buffer{},
	})
	if code == 0 && err == nil {
		t.Fatal("expected policy block for destructive alias expansion")
	}
	if !strings.Contains(stderr.String(), "blocked by policy") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestE2EAliasGitStatusExplain(t *testing.T) {
	setupCLXHomeForHost(t, nil)

	ctx := context.Background()
	store, err := aliases.Open(ctx, 32)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(ctx, "gst", "git status"); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	code, err := pipeline.Run(ctx, config.Default(), "gst", pipeline.Options{
		Explain:    true,
		AliasStore: store,
		Stdout:     &stdout,
		Stderr:     &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stdout.String(), "git") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

// fakeCmdProvider is shared with pipeline aicommand tests pattern.
type fakeCmdProvider struct {
	resp *providers.CommandResponse
	err  error
}

func (f *fakeCmdProvider) Name() string { return "fake" }

func (f *fakeCmdProvider) ResolveIntent(context.Context, providers.IntentRequest) (*providers.IntentResponse, error) {
	return nil, providers.ErrNoMatch
}

func (f *fakeCmdProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return "", nil
}

func (f *fakeCmdProvider) GenerateCommand(context.Context, providers.CommandRequest) (*providers.CommandResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}
