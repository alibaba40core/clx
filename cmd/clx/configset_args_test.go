package main

import "testing"

func TestParseConfigSetArgsFlagsAfterPath(t *testing.T) {
	t.Parallel()
	path, useStdin, pos, errMsg := parseConfigSetArgs([]string{
		"providers.openai.api_key", "--stdin",
	})
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	if !useStdin {
		t.Fatal("expected useStdin")
	}
	if len(pos) != 1 || pos[0] != "providers.openai.api_key" {
		t.Fatalf("positional=%v", pos)
	}
	if path != "" {
		t.Fatalf("config path=%q", path)
	}
}

func TestParseConfigSetArgsConfigAnywhere(t *testing.T) {
	t.Parallel()
	path, useStdin, pos, errMsg := parseConfigSetArgs([]string{
		"--config", "/tmp/cfg.yaml", "provider", "openai",
	})
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	if useStdin {
		t.Fatal("unexpected stdin")
	}
	if path != "/tmp/cfg.yaml" {
		t.Fatalf("path=%q", path)
	}
	if len(pos) != 2 {
		t.Fatalf("positional=%v", pos)
	}
}
