package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Save validates cfg, encrypts secrets, and atomically writes config.yaml.
func Save(ctx context.Context, path string, cfg Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := Validate(cfg); err != nil {
		return err
	}
	disk, err := PrepareForDisk(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".clx-config-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}
	if err := Encode(disk, tmp); err != nil {
		cleanup()
		return fmt.Errorf("encode config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpPath, filePerm); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}
	return nil
}
