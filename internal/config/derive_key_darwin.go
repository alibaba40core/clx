//go:build darwin

package config

import (
	"os"
	"strings"
)

// machineIdentity uses hostname + user on macOS (no subprocess / CGO for IOPlatformUUID).
func machineIdentity() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return "", os.ErrInvalid
	}
	user := os.Getenv("USER")
	if user == "" {
		return "", os.ErrInvalid
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return host + "\x00" + user + "\x00" + home, nil
}
