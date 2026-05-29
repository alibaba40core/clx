package config

import (
	"crypto/sha256"
)

const keyDerivationSalt = "clx-secrets-v1"

func deriveMasterKey() ([]byte, error) {
	identity, err := machineIdentity()
	if err == nil && identity != "" {
		sum := sha256.Sum256([]byte(keyDerivationSalt + "\x00" + identity))
		return sum[:], nil
	}
	return loadOrCreateFallbackKey()
}
