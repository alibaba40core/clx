package config

// DefaultOllamaModel is the shipped default model for the Ollama provider.
//
// Picked for: Apache 2.0, native JSON/schema support, ~1.4 GB on disk,
// correct closed-vocabulary intent + param extraction on CPU (Phase 2.1).
// Local benchmark (scripts/bench-ollama-models.ps1, "show current directory"):
//   gemma3:270m ~13s but wrong intent; qwen3:1.7b ~67s correct; qwen3:4b ~179s correct.
//
// Tested alternates (swap in config.yaml under providers.ollama.model):
//   - qwen3:4b     — quality bump when GPU/RAM allows (~2.5 GB, ~3x slower on CPU)
//   - qwen2.5:7b   — higher quality; ~4.4 GB
//   - llama3.1:8b  — Meta-ecosystem alternate; avoid plain llama3 (weak tool JSON)
const DefaultOllamaModel = "qwen3:1.7b"

// DefaultGeminiModel is the shipped default model for the Gemini provider.
// gemini-2.0-flash: fast, supports structured output (responseSchema), good at
// classification tasks. Free tier limited to 15 RPM.
const DefaultGeminiModel = "gemini-2.0-flash"

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
			Gemini: GeminiProvider{
				APIKey: "",
				Model:  DefaultGeminiModel,
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
			DryRun:              false,
		},
		Features: FeaturesConfig{
			Explain:             true,
			CacheCommands:       true,
			LearningMode:        false,
			AICommandGeneration: true,
		},
		Cache: CacheConfig{
			MaxEntries:   1024,
			TTLDays:      30,
			MaxDiskBytes: 5 * 1024 * 1024,
		},
		Logging: LoggingConfig{
			Enabled: true,
			Level:   "info",
		},
	}
}
