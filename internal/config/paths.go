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

// MemoryDir returns ~/.clx/memory.
func MemoryDir() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "memory"), nil
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
