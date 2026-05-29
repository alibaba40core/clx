package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestReadPlainStdin(t *testing.T) {
	t.Parallel()
	val, err := readPlainStdin(strings.NewReader("  hello world \n"))
	if err != nil {
		t.Fatal(err)
	}
	if val != "hello world" {
		t.Fatalf("got %q", val)
	}
}

func TestReadSecretValueFromPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = w.WriteString("sk-piped-secret\n")
		_ = w.Close()
	}()

	var stderr bytes.Buffer
	val, err := readSecretValue(r, &stderr, "providers.openai.api_key")
	if err != nil {
		t.Fatal(err)
	}
	if val != "sk-piped-secret" {
		t.Fatalf("got %q", val)
	}
}

func TestReadSecretValueEmptyPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	var stderr bytes.Buffer
	_, err = readSecretValue(r, &stderr, "providers.openai.api_key")
	if err == nil {
		t.Fatal("expected error for empty pipe")
	}
}
