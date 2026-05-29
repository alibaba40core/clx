package config

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

const fallbackKeyBytes = 32

func loadOrCreateFallbackKey() ([]byte, error) {
	path, err := SecretKeyPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err == nil {
		if len(data) != fallbackKeyBytes {
			return nil, fmt.Errorf("invalid fallback key size at %s", path)
		}
		return data, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	key := make([]byte, fallbackKeyBytes)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate fallback key: %w", err)
	}
	home, err := Home()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(home, dirPerm); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(home, ".secret-key-*")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}
	if _, err := tmp.Write(key); err != nil {
		cleanup()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return nil, err
	}
	if err := os.Chmod(tmpPath, filePerm); err != nil {
		cleanup()
		return nil, err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return nil, err
	}
	return key, nil
}
