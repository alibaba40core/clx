package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	plain := "sk-test-secret-key-12345"
	enc, err := EncryptSecret(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !IsEncrypted(enc) {
		t.Fatalf("expected enc prefix, got %q", enc)
	}
	got, err := DecryptSecret(enc)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestPrepareForDiskEncryptsPlaintext(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	cfg := Default()
	cfg.Providers.OpenAI.APIKey = "sk-plaintext-key"
	disk, err := PrepareForDisk(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !IsEncrypted(disk.Providers.OpenAI.APIKey) {
		t.Fatalf("expected encrypted on disk, got %q", disk.Providers.OpenAI.APIKey)
	}
	if cfg.Providers.OpenAI.APIKey != "sk-plaintext-key" {
		t.Fatal("PrepareForDisk must not mutate input")
	}
}

func TestSaveLoadEncryptsAPIKey(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg := Default()
	cfg.Providers.OpenAI.APIKey = "sk-save-test-key"
	if err := Save(context.Background(), path, cfg); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "sk-save-test-key") {
		t.Fatalf("plaintext key in config file: %s", raw)
	}
	if !strings.Contains(string(raw), "enc:v1:") {
		t.Fatalf("expected enc:v1 in file: %s", raw)
	}

	loaded, err := Load(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Providers.OpenAI.APIKey != "sk-save-test-key" {
		t.Fatalf("loaded key = %q", loaded.Providers.OpenAI.APIKey)
	}
}
