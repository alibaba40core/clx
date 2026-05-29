package config

// Config is the root CLX runtime configuration (see configs/config.example.yaml).
type Config struct {
	Provider  string          `yaml:"provider"`
	Model     string          `yaml:"model"`
	Providers ProvidersConfig `yaml:"providers"`
	Execution ExecutionConfig `yaml:"execution"`
	Safety    SafetyConfig    `yaml:"safety"`
	Features  FeaturesConfig  `yaml:"features"`
	Cache     CacheConfig     `yaml:"cache"`
	Logging   LoggingConfig   `yaml:"logging"`
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
	AutoExecute      bool `yaml:"auto_execute"`
	Timeout          int  `yaml:"timeout"`
	ShellIntegration bool `yaml:"shell_integration"`
}

// SafetyConfig controls safety defaults.
//
// Mode is the user-facing safety preset. Currently informational; Phase 3
// will wire it to behavior bundles:
//
//   - low: execute directly, no dry-run, no confirm (risk-engine High-risk
//     override still applies for destructive commands)
//   - medium: dry-run preview, then y/n confirm before exec (default)
//   - high: dry-run preview plus verbose display; refuses -y shortcut;
//     always requires explicit confirm
//
// DryRun and RequireConfirmation remain explicit overrides; when set, they
// win over the Mode-implied default.
type SafetyConfig struct {
	Mode                string `yaml:"mode"`
	RequireConfirmation bool   `yaml:"require_confirmation"`
	DryRun              bool   `yaml:"dry_run"`
}

// FeaturesConfig toggles optional features.
type FeaturesConfig struct {
	Explain       bool `yaml:"explain"`
	CacheCommands bool `yaml:"cache_commands"`
	LearningMode  bool `yaml:"learning_mode"`
}

// CacheConfig bounds the file-backed intent cache (~/.clx/cache/intents.json).
type CacheConfig struct {
	MaxEntries   int `yaml:"max_entries"`
	TTLDays      int `yaml:"ttl_days"`
	MaxDiskBytes int `yaml:"max_disk_bytes"`
}

// LoggingConfig controls structured logging.
type LoggingConfig struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
}
