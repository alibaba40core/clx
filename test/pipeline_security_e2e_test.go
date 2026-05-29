package test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/pipeline"
	"github.com/alibaba40core/clx/internal/providers"
)

func TestE2EExplainAIEnrichment(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	stub := &explainE2EProvider{text: "AI-generated explanation for listing."}
	var stdout bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "unknown phrase xyz", pipeline.Options{
		Explain:  true,
		Provider: stub,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "list_dir", Params: map[string]string{}, Confidence: 0.9,
			Source: intent.SourceAI,
		}},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stdout.String(), "AI-generated explanation") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestE2ESecurityMaliciousIntentNeverExecs(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	cfg := config.Default()
	cfg.Safety.DryRun = false
	cfg.Execution.AutoExecute = true
	cfg.Safety.RequireConfirmation = false

	var stderr bytes.Buffer
	code, err := pipeline.Run(context.Background(), cfg, "unknown xyz", pipeline.Options{
		Yes: true,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "rm_rf_slash", Params: map[string]string{}, Source: intent.SourceAI,
		}},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !strings.Contains(stderr.String(), "untrusted resolver output rejected") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestE2ESecurityMaliciousExplainNeverInArgv(t *testing.T) {
	setupCLXHomeForHost(t, nil)
	evil := "; rm -rf /"
	stub := &explainE2EProvider{text: evil}
	var stdout bytes.Buffer
	code, err := pipeline.Run(context.Background(), config.Default(), "unknown xyz", pipeline.Options{
		Explain:  true,
		Provider: stub,
		AIResolver: stubAIResult{result: intent.ResolvedIntent{
			Intent: "list_dir", Source: intent.SourceAI, Confidence: 0.9,
		}},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	out := stdout.String()
	if !strings.Contains(out, evil) {
		t.Fatalf("explain text should appear in display: %q", out)
	}
	cmdLine := commandLine(out)
	if strings.Contains(cmdLine, "rm -rf") {
		t.Fatalf("malicious explain leaked into command line: %q", cmdLine)
	}
}

type explainE2EProvider struct {
	text string
}

func (explainE2EProvider) Name() string { return "stub" }

func (explainE2EProvider) ResolveIntent(context.Context, providers.IntentRequest) (*providers.IntentResponse, error) {
	return nil, providers.ErrNoMatch
}

func (e explainE2EProvider) Explain(context.Context, generator.GeneratedCommand) (string, error) {
	return e.text, nil
}

var _ providers.Provider = explainE2EProvider{}
