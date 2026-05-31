package pipeline

import (
	"bytes"
	"context"
	"sync/atomic"
	"testing"

	"github.com/alibaba40core/clx/internal/cache"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/policy"
	"github.com/alibaba40core/clx/internal/providers"
)

// fakeCmdProvider implements both providers.Provider and providers.CommandGenerator.
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

type countingCmdProvider struct {
	fakeCmdProvider
	calls atomic.Int32
}

func (c *countingCmdProvider) GenerateCommand(ctx context.Context, req providers.CommandRequest) (*providers.CommandResponse, error) {
	c.calls.Add(1)
	return c.fakeCmdProvider.GenerateCommand(ctx, req)
}

func newAICmdEnv(t *testing.T) {
	t.Helper()
	t.Setenv("CLX_HOME", t.TempDir())
	_, _ = config.Bootstrap(context.Background())
	policy.ResetCache()
	testProfile(t, "linux", "bash")
}

func TestRunAICommandFallbackDryRun(t *testing.T) {
	newAICmdEnv(t)

	var stdout, stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "totally unknown phrase here", Options{
		DryRun:     true,
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider: &fakeCmdProvider{resp: &providers.CommandResponse{
			Argv:        []string{"ls", "-la"},
			Shell:       "bash",
			Explanation: "list all files",
			Confidence:  0.9,
		}},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"Source:      AI", "ai-generated command", "ls", "dry-run:"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("stdout missing %q\n%s", want, out)
		}
	}
}

func TestRunAICommandRejectsMaliciousArgv(t *testing.T) {
	newAICmdEnv(t)

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "totally unknown phrase here", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider: &fakeCmdProvider{resp: &providers.CommandResponse{
			Argv:       []string{"ls", "|", "sh"},
			Shell:      "bash",
			Confidence: 0.95,
		}},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("expected rejection, code=%d err=%v", code, err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("rejected as unsafe")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunAICommandLowConfidence(t *testing.T) {
	newAICmdEnv(t)

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "totally unknown phrase here", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider: &fakeCmdProvider{resp: &providers.CommandResponse{
			Argv:       []string{"ls"},
			Shell:      "bash",
			Confidence: 0.1,
		}},
		Stdout: &bytes.Buffer{},
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("expected low-confidence drop, code=%d err=%v", code, err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("not confident enough")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunAICommandFeatureDisabled(t *testing.T) {
	newAICmdEnv(t)

	cfg := config.Default()
	cfg.Features.AICommandGeneration = false

	var stdout, stderr bytes.Buffer
	code, err := Run(context.Background(), cfg, "totally unknown phrase here", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider: &fakeCmdProvider{resp: &providers.CommandResponse{
			Argv:       []string{"ls"},
			Shell:      "bash",
			Confidence: 0.95,
		}},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("expected miss, code=%d err=%v", code, err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("AI could not map")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
	if bytes.Contains(stdout.Bytes(), []byte("ai-generated")) {
		t.Fatalf("should not have generated a command: %s", stdout.String())
	}
}

func TestRunAICommandProviderRateLimited(t *testing.T) {
	newAICmdEnv(t)

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "totally unknown phrase here", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider:   &fakeCmdProvider{err: providers.ErrRateLimited},
		Stdout:     &bytes.Buffer{},
		Stderr:     &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("expected failure, code=%d err=%v", code, err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("rate limit")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
	if bytes.Contains(stderr.Bytes(), []byte("try rephrasing")) {
		t.Fatalf("stderr should not say rephrasing: %q", stderr.String())
	}
}

func TestRunAICommandCacheSkipsSecondProviderCall(t *testing.T) {
	newAICmdEnv(t)

	cmdPath, err := config.CacheCommandsPath()
	if err != nil {
		t.Fatal(err)
	}
	cmdStore, err := cache.LoadCommands(context.Background(), cmdPath, config.Default().Cache, nil)
	if err != nil {
		t.Fatal(err)
	}

	prov := &countingCmdProvider{fakeCmdProvider: fakeCmdProvider{resp: &providers.CommandResponse{
		Argv:        []string{"ls", "-la"},
		Shell:       "bash",
		Explanation: "list all files",
		Confidence:  0.9,
	}}}
	opts := Options{
		DryRun:       true,
		AIResolver:   &fakeAIResolver{err: intent.ErrNotFound},
		Provider:     prov,
		CommandCache: cmdStore,
		Stdout:       &bytes.Buffer{},
		Stderr:       &bytes.Buffer{},
	}
	raw := "totally unknown phrase here"
	cfg := config.Default()

	code, err := Run(context.Background(), cfg, raw, opts)
	if err != nil || code != 0 {
		t.Fatalf("first run: code=%d err=%v", code, err)
	}
	if prov.calls.Load() != 1 {
		t.Fatalf("first run calls=%d want 1", prov.calls.Load())
	}

	code, err = Run(context.Background(), cfg, raw, opts)
	if err != nil || code != 0 {
		t.Fatalf("second run: code=%d err=%v", code, err)
	}
	if prov.calls.Load() != 1 {
		t.Fatalf("second run should use cache, calls=%d want 1", prov.calls.Load())
	}
}

func TestRunAICommandProviderUnavailable(t *testing.T) {
	newAICmdEnv(t)

	var stderr bytes.Buffer
	code, err := Run(context.Background(), config.Default(), "totally unknown phrase here", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider:   &fakeCmdProvider{err: providers.ErrUnavailable},
		Stdout:     &bytes.Buffer{},
		Stderr:     &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("expected failure, code=%d err=%v", code, err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("AI provider unavailable")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunAICommandOllamaUnavailableWSLHint(t *testing.T) {
	newAICmdEnv(t)

	cfg := config.Default()
	cfg.Provider = "ollama"
	cfg.Providers.Ollama.Host = "http://localhost:11434"

	var stderr bytes.Buffer
	code, err := Run(context.Background(), cfg, "totally unknown phrase here", Options{
		AIResolver: &fakeAIResolver{err: intent.ErrNotFound},
		Provider:   &fakeCmdProvider{err: providers.ErrUnavailable},
		Stdout:     &bytes.Buffer{},
		Stderr:     &stderr,
	})
	if code != 1 || err == nil {
		t.Fatalf("expected failure, code=%d err=%v", code, err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Ollama runs in WSL")) {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
