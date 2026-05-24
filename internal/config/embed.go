package config

import _ "embed"

//go:embed templates/config.yaml
var embeddedConfigYAML []byte

//go:embed templates/policy.yaml
var embeddedPolicyYAML []byte

// EmbeddedConfigYAML returns the default config template bytes.
func EmbeddedConfigYAML() []byte {
	return embeddedConfigYAML
}

// EmbeddedPolicyYAML returns the default policy template bytes.
func EmbeddedPolicyYAML() []byte {
	return embeddedPolicyYAML
}
