//go:build lite

package pipeline

import (
	"testing"

	"github.com/alibaba40core/clx/internal/config"
)

func TestNeedsAIWorker(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want bool
	}{
		{
			name: "provider none",
			cfg: config.Config{
				Provider: "none",
				Providers: config.ProvidersConfig{Primary: "none"},
			},
			want: false,
		},
		{
			name: "ollama enabled",
			cfg: config.Config{
				Provider:  "ollama",
				Providers: config.ProvidersConfig{Primary: "ollama"},
			},
			want: true,
		},
		{
			name: "ai command generation only",
			cfg: config.Config{
				Provider: "none",
				Features: config.FeaturesConfig{AICommandGeneration: true},
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := needsAIWorker(tc.cfg); got != tc.want {
				t.Fatalf("needsAIWorker() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAiEnabled(t *testing.T) {
	cfg := config.Config{Provider: "openai", Providers: config.ProvidersConfig{Primary: "openai"}}
	if !aiEnabled(cfg) {
		t.Fatal("expected ai enabled")
	}
	cfg.Provider = "none"
	cfg.Providers.Primary = "none"
	if aiEnabled(cfg) {
		t.Fatal("expected ai disabled for none")
	}
}
