package config

import "fmt"

// DecryptConfig decrypts enc:v1 API key fields into plaintext for runtime use.
func DecryptConfig(cfg *Config) error {
	if err := decryptField(&cfg.Providers.OpenAI.APIKey, "providers.openai.api_key"); err != nil {
		return err
	}
	if err := decryptField(&cfg.Providers.Azure.APIKey, "providers.azure.api_key"); err != nil {
		return err
	}
	return nil
}

func decryptField(field *string, name string) error {
	if *field == "" || !IsEncrypted(*field) {
		return nil
	}
	plain, err := DecryptSecret(*field)
	if err != nil {
		return fmt.Errorf("decrypt %s: %w", name, err)
	}
	*field = plain
	return nil
}

// PrepareForDisk returns a copy of cfg with secret fields encrypted for persistence.
func PrepareForDisk(cfg Config) (Config, error) {
	out := cfg
	var err error
	if out.Providers.OpenAI.APIKey != "" && !IsEncrypted(out.Providers.OpenAI.APIKey) {
		out.Providers.OpenAI.APIKey, err = EncryptSecret(out.Providers.OpenAI.APIKey)
		if err != nil {
			return Config{}, fmt.Errorf("encrypt providers.openai.api_key: %w", err)
		}
	}
	if out.Providers.Azure.APIKey != "" && !IsEncrypted(out.Providers.Azure.APIKey) {
		out.Providers.Azure.APIKey, err = EncryptSecret(out.Providers.Azure.APIKey)
		if err != nil {
			return Config{}, fmt.Errorf("encrypt providers.azure.api_key: %w", err)
		}
	}
	return out, nil
}

// EncryptSecretsOnDisk re-encrypts any plaintext secret fields in cfg (in-place on disk copy).
func EncryptSecretsOnDisk(cfg *Config) error {
	prepared, err := PrepareForDisk(*cfg)
	if err != nil {
		return err
	}
	cfg.Providers.OpenAI.APIKey = prepared.Providers.OpenAI.APIKey
	cfg.Providers.Azure.APIKey = prepared.Providers.Azure.APIKey
	return nil
}

// MaskSecret masks a secret value for display (never prints full plaintext).
func MaskSecret(value string) string {
	if value == "" {
		return `""`
	}
	if len(value) >= 4 {
		return "****" + value[len(value)-4:]
	}
	return "<set>"
}

// IsSecretPath reports whether path holds a credential field.
func IsSecretPath(path string) bool {
	switch path {
	case "providers.openai.api_key", "providers.azure.api_key":
		return true
	default:
		return false
	}
}
