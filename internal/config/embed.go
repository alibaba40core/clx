package config

import _ "embed"

//go:embed templates/config.yaml
var embeddedConfigYAML []byte

//go:embed templates/policy.yaml
var embeddedPolicyYAML []byte

//go:embed templates/aliases.yaml
var embeddedAliasesYAML []byte

//go:embed templates/shell-integration.txt
var embeddedShellIntegration []byte

// EmbeddedConfigYAML returns the default config template bytes.
func EmbeddedConfigYAML() []byte {
	return embeddedConfigYAML
}

// EmbeddedPolicyYAML returns the default policy template bytes.
func EmbeddedPolicyYAML() []byte {
	return embeddedPolicyYAML
}

// EmbeddedAliasesYAML returns the default aliases template bytes.
func EmbeddedAliasesYAML() []byte {
	return embeddedAliasesYAML
}

// EmbeddedShellIntegration returns install instructions for shell hooks.
func EmbeddedShellIntegration() []byte {
	return embeddedShellIntegration
}
