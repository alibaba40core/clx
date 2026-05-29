//go:build windows

package config

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func machineIdentity() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	guid, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", err
	}
	user := os.Getenv("USERNAME")
	if user == "" {
		user, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(guid) + "\x00" + user, nil
}
