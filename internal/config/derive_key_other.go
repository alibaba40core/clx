//go:build !windows && !linux && !darwin

package config

import "errors"

var errNoMachineIdentity = errors.New("machine identity unavailable")

func machineIdentity() (string, error) {
	return "", errNoMachineIdentity
}
