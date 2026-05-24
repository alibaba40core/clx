package config

// Config is the root CLX runtime configuration (see configs/config.example.yaml).
type Config struct {
	Provider  string           `yaml:"provider"`
	Model     string           `yaml:"model"`
	Providers ProvidersConfig  `yaml:"providers"`
	Execution ExecutionConfig  `yaml:"execution"`
	Safety    SafetyConfig     `yaml:"safety"`
	Features  FeaturesConfig   `yaml:"features"`
	Logging   LoggingConfig    `yaml:"logging"`
}

// ProvidersConfig holds per-provider settings.
type ProvidersConfig struct {
	Ollama OllamaProvider `yaml:"ollama"`
	OpenAI OpenAIProvider `yaml:"openai"`
	Azure  AzureProvider  `yaml:"azure"`
}

// OllamaProvider configures the local Ollama provider.
type OllamaProvider struct {
	Host  string `yaml:"host"`
	Model string `yaml:"model"`
}

// OpenAIProvider configures the OpenAI provider.
type OpenAIProvider struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// AzureProvider configures the Azure AI provider.
type AzureProvider struct {
	Endpoint   string `yaml:"endpoint"`
	APIKey     string `yaml:"api_key"`
	Deployment string `yaml:"deployment"`
}

// ExecutionConfig controls command execution behavior.
type ExecutionConfig struct {
	AutoExecute       bool `yaml:"auto_execute"`
	Timeout           int  `yaml:"timeout"`
	ShellIntegration  bool `yaml:"shell_integration"`
}

// SafetyConfig controls safety defaults (enforced in Phase 3).
type SafetyConfig struct {
	Level               string `yaml:"level"`
	RequireConfirmation bool   `yaml:"require_confirmation"`
	DryRun              bool   `yaml:"dry_run"`
}

// FeaturesConfig toggles optional features.
type FeaturesConfig struct {
	Explain       bool `yaml:"explain"`
	CacheCommands bool `yaml:"cache_commands"`
	LearningMode  bool `yaml:"learning_mode"`
}

// LoggingConfig controls structured logging.
type LoggingConfig struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
}
