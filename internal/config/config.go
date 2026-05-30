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
	Aliases   AliasesConfig   `yaml:"aliases"`
	Memory    MemoryConfig    `yaml:"memory"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// MemoryConfig bounds session-scoped memory files.
type MemoryConfig struct {
	Enabled             bool `yaml:"enabled"`
	MaxEntriesPerSession int `yaml:"max_entries_per_session"`
	MaxSessions         int  `yaml:"max_sessions"`
	TTLDays             int  `yaml:"ttl_days"`
}

// AliasesConfig bounds user-global alias shortcuts.
type AliasesConfig struct {
	MaxAliases int `yaml:"max_aliases"`
}

// ProvidersConfig holds per-provider settings and optional fallback chain (D6).
type ProvidersConfig struct {
	Primary  string         `yaml:"primary"`
	Fallback string         `yaml:"fallback"`
	Timeout  int            `yaml:"timeout"` // seconds; 0 = use execution.timeout
	Ollama   OllamaProvider `yaml:"ollama"`
	OpenAI   OpenAIProvider `yaml:"openai"`
	Azure    AzureProvider  `yaml:"azure"`
	Gemini   GeminiProvider `yaml:"gemini"`
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

// GeminiProvider configures the Google Gemini provider.
type GeminiProvider struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// ExecutionConfig controls command execution behavior.
type ExecutionConfig struct {
	AutoExecute      bool `yaml:"auto_execute"`
	Timeout          int  `yaml:"timeout"`
	ShellIntegration bool `yaml:"shell_integration"`
}

// SafetyConfig controls how commands are gated after risk classification.
//
// Mode selects a preset matrix (see DecideSafetyAction in safety.go):
//
//   - low:    low/medium risk run; high risk confirm
//   - medium: low risk run; medium/high explain + confirm (default)
//   - high:   low explain + confirm; medium/high preview + explain + confirm;
//             -y cannot skip confirm for medium/high risk
//   - custom: RequireConfirmation, DryRun, and Features.Explain apply globally
//
// RequireConfirmation and DryRun only drive behavior when Mode is custom.
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
	// AICommandGeneration enables the hybrid fallback where, after rules and
	// cache miss, a configured AI provider generates a full command (argv).
	// The generated argv is still validated, risk-assessed, policy-gated, and
	// confirmed before exec. Default on (see config.Default).
	AICommandGeneration bool `yaml:"ai_command_generation"`
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
