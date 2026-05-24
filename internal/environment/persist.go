package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxProfileBytes = 64 * 1024

// Load reads a system profile from path.
func Load(ctx context.Context, path string) (SystemProfile, error) {
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return SystemProfile{}, err
	}
	defer f.Close()

	dec := json.NewDecoder(io.LimitReader(f, maxProfileBytes))
	var p SystemProfile
	if err := dec.Decode(&p); err != nil {
		return SystemProfile{}, fmt.Errorf("decode profile: %w", err)
	}
	return p, nil
}

// Save writes a system profile to path atomically.
func Save(ctx context.Context, path string, p SystemProfile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.SchemaVersion = SchemaVersion

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".clx-profile-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}
	return nil
}
