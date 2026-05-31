package config

import (
	"os"
	"path/filepath"
)

const homeDirName = ".clx"

// Home returns the CLX runtime home directory (~/.clx or $CLX_HOME).
func Home() (string, error) {
	if v := os.Getenv("CLX_HOME"); v != "" {
		return filepath.Clean(v), nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, homeDirName), nil
}

// SecretKeyPath returns ~/.clx/.secret-key (fallback encryption key material).
func SecretKeyPath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".secret-key"), nil
}

// ConfigPath returns the path to config.yaml under the CLX home.
func ConfigPath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config.yaml"), nil
}

// LogsDir returns ~/.clx/logs.
func LogsDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "logs"), nil
}

// SessionsDir returns ~/.clx/sessions.
func SessionsDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "sessions"), nil
}

// PoliciesDir returns ~/.clx/policies.
func PoliciesDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "policies"), nil
}

// AliasesPath returns ~/.clx/aliases.yaml.
func AliasesPath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "aliases.yaml"), nil
}

// PolicyPath returns ~/.clx/policies/policy.yaml.
func PolicyPath() (string, error) {
	dir, err := PoliciesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "policy.yaml"), nil
}

// CacheDir returns ~/.clx/cache.
func CacheDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache"), nil
}

// CacheIntentsPath returns ~/.clx/cache/intents.json.
func CacheIntentsPath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "intents.json"), nil
}

// CacheExplanationsPath returns ~/.clx/cache/explanations.json.
func CacheExplanationsPath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "explanations.json"), nil
}

// CacheCommandsPath returns ~/.clx/cache/commands.json.
func CacheCommandsPath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "commands.json"), nil
}

// MemoryDir returns ~/.clx/memory.
func MemoryDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "memory"), nil
}

// UserRulesDir returns ~/.clx/rules (user rule overrides).
func UserRulesDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "rules"), nil
}

// UserSkillsDir returns ~/.clx/skills (user skill pack overrides).
func UserSkillsDir() (string, error) {
	return SkillsDir()
}

// SkillsDir returns ~/.clx/skills (user skill overrides).
func SkillsDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "skills"), nil
}

// SystemProfilePath returns ~/.clx/system_profile.json.
func SystemProfilePath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "system_profile.json"), nil
}
