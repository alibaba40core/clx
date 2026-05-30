package test

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/memory"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/pipeline"
	"github.com/alibaba40core/clx/internal/policy"
)

type memoryStubAI struct{}

func (memoryStubAI) Resolve(context.Context, parser.Request) (intent.ResolvedIntent, error) {
	return intent.ResolvedIntent{}, intent.ErrNotFound
}

func TestE2EMemoryFollowUpWithoutAI(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux profile for grep intent")
	}
	home := t.TempDir()
	t.Setenv("CLX_HOME", home)
	t.Setenv("CLX_SESSION_ID", "e2e-mem")
	ctx := context.Background()
	if _, err := config.Bootstrap(ctx); err != nil {
		t.Fatal(err)
	}
	policy.ResetCache()
	path, err := config.SystemProfilePath()
	if err != nil {
		t.Fatal(err)
	}
	store := environment.NewProfileStore()
	store.UpsertProfile(environment.SystemProfile{
		OS:             "linux",
		Shell:          "bash",
		AvailableTools: []string{"grep"},
	})
	if err := environment.SaveStore(ctx, path, store); err != nil {
		t.Fatal(err)
	}

	memStore, err := memory.Open(ctx, "e2e-mem", config.Default().Memory)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	var stdout, stderr bytes.Buffer
	code, err := pipeline.Run(ctx, cfg, "grep errors logs.txt", pipeline.Options{
		Explain:     true,
		MemoryStore: memStore,
		AIResolver:  memoryStubAI{},
		Stdout:      &stdout,
		Stderr:      &stderr,
	})
	if err != nil || code != 0 {
		t.Fatalf("first run code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "search_text_in_file") {
		t.Fatalf("first stdout=%q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code, err = pipeline.Run(ctx, cfg, "again", pipeline.Options{
		Explain:     true,
		MemoryStore: memStore,
		AIResolver:  memoryStubAI{},
		Stdout:      &stdout,
		Stderr:      &stderr,
	})
	if err != nil || code != 0 {
		t.Fatalf("follow-up code=%d err=%v stderr=%s", code, err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "search_text_in_file") {
		t.Fatalf("follow-up stdout=%q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "grep") {
		t.Fatalf("follow-up should render grep command, stdout=%q", stdout.String())
	}
}
