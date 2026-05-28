package config

// DefaultOllamaModel is the shipped default model for the Ollama provider.
//
// Picked for: Apache 2.0, native tool/JSON support, ~3 GB RAM at Q4,
// strong on closed-vocabulary intent classification (Phase 2.1 use case).
//
// Tested alternates (swap in config.yaml under providers.ollama.model):
//   - qwen3:1.7b   — lightest viable; relies on Ollama structured outputs
//   - qwen2.5:7b   — quality bump; ~4.4 GB, 8/10 tool-use score
//   - llama3.1:8b  — Meta-ecosystem alternate; ~4.7 GB, reliable workhorse
const DefaultOllamaModel = "qwen3:4b"

// Default returns the built-in default configuration.
func Default() Config {
	return Config{
		Provider: "ollama",
		Model:    DefaultOllamaModel,
		Providers: ProvidersConfig{
			Ollama: OllamaProvider{
				Host:  "http://localhost:11434",
				Model: DefaultOllamaModel,
			},
			OpenAI: OpenAIProvider{
				APIKey: "",
				Model:  "gpt-4.1-mini",
			},
			Azure: AzureProvider{
				Endpoint:   "",
				APIKey:     "",
				Deployment: "",
			},
		},
		Execution: ExecutionConfig{
			AutoExecute:      false,
			Timeout:          30,
			ShellIntegration: false,
		},
		Safety: SafetyConfig{
			Mode:                "medium",
			RequireConfirmation: true,
			DryRun:              true,
		},
		Features: FeaturesConfig{
			Explain:       true,
			CacheCommands: true,
			LearningMode:  false,
		},
		Logging: LoggingConfig{
			Enabled: true,
			Level:   "info",
		},
	}
}
