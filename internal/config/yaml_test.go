package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeExampleConfig(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "configs", "config.example.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		// Fall back to embedded template.
		data = EmbeddedConfigYAML()
	}
	root, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got, ok := root.get("provider"); !ok || got != "ollama" {
		t.Errorf("provider = %q, ok=%v", got, ok)
	}
	if got, ok := root.get("execution", "timeout"); !ok || got != "30" {
		t.Errorf("execution.timeout = %q", got)
	}
}

func TestDecodeRejectsLists(t *testing.T) {
	t.Parallel()
	_, err := Decode(strings.NewReader("blocked:\n  - item\n"))
	if err == nil || !strings.Contains(err.Error(), "lists") {
		t.Fatalf("expected list error, got %v", err)
	}
}

func TestDecodeRejectsFlowSyntax(t *testing.T) {
	t.Parallel()
	_, err := Decode(strings.NewReader("key: {a: 1}\n"))
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
	root, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	applyNode(&cfg, root)
	if cfg.Provider != def.Provider || cfg.Execution.Timeout != def.Execution.Timeout {
		t.Errorf("round trip mismatch: %+v", cfg)
	}
}

func FuzzDecodeConfigSubset(f *testing.F) {
	f.Add([]byte("provider: ollama\n"))
	f.Add([]byte("execution:\n  timeout: 30\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxYAMLBytes {
			return
		}
		_, _ = Decode(bytes.NewReader(data))
	})
}
