package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

func TestDecodeExampleConfig(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "configs", "config.example.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		data = EmbeddedConfigYAML()
	}
	root, err := yamlutil.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got, ok := root.GetString("provider"); !ok || got != "ollama" {
		t.Errorf("provider = %q, ok=%v", got, ok)
	}
	if got, ok := root.GetString("execution", "timeout"); !ok || got != "30" {
		t.Errorf("execution.timeout = %q", got)
	}
}

func TestDecodeRejectsFlowSyntax(t *testing.T) {
	t.Parallel()
	_, err := yamlutil.Decode(strings.NewReader(`key: {"a": 1}` + "\n"))
	if err == nil {
		t.Fatal("expected error for flow syntax")
	}
}

func TestEncodeRoundTrip(t *testing.T) {
	t.Parallel()
	def := Default()
	var buf bytes.Buffer
	if err := Encode(def, &buf); err != nil {
		t.Fatal(err)
	}
	root, err := yamlutil.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	applyNode(&cfg, root)
	if cfg.Provider != def.Provider || cfg.Execution.Timeout != def.Execution.Timeout {
		t.Errorf("round trip mismatch: %+v", cfg)
	}
}
