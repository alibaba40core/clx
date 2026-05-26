package config

// Default returns the built-in default configuration.
func Default() Config {
	return Config{
		Provider: "ollama",
		Model:    "llama3",
		Providers: ProvidersConfig{
			Ollama: OllamaProvider{
				Host:  "http://localhost:11434",
				Model: "llama3",
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
