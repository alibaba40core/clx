//go:build linux

package config

import (
	"os"
	"strings"
)

func machineIdentity() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", os.ErrInvalid
	}
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("LOGNAME")
	}
	if user == "" {
		return "", os.ErrInvalid
	}
	return id + "\x00" + user, nil
}
