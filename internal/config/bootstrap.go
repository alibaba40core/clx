package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const dirPerm = 0o700
const filePerm = 0o600

var bootstrapDirs = []func() (string, error){
	CacheDir,
	MemoryDir,
	SessionsDir,
	PoliciesDir,
	SkillsDir,
	LogsDir,
}

// BootstrapResult reports what first-run bootstrap did.
type BootstrapResult struct {
	CreatedDirs []string
	WroteConfig bool
	WrotePolicy bool
	WroteProfile bool
}

// Bootstrap ensures ~/.clx/ exists with default files (idempotent).
func Bootstrap(ctx context.Context) (BootstrapResult, error) {
	if err := ctx.Err(); err != nil {
		return BootstrapResult{}, err
	}

	var result BootstrapResult

	home, err := Home()
	if err != nil {
		return BootstrapResult{}, err
	}

	if err := os.MkdirAll(home, dirPerm); err != nil {
		return BootstrapResult{}, fmt.Errorf("create home %s: %w", home, err)
	}

	for _, dirFn := range bootstrapDirs {
		if err := ctx.Err(); err != nil {
			return BootstrapResult{}, err
		}
		dir, err := dirFn()
		if err != nil {
			return BootstrapResult{}, err
		}
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return BootstrapResult{}, fmt.Errorf("create dir %s: %w", dir, err)
		}
		result.CreatedDirs = append(result.CreatedDirs, dir)
	}

	cfgPath, err := ConfigPath()
	if err != nil {
		return BootstrapResult{}, err
	}
	wrote, err := writeIfMissing(ctx, cfgPath, EmbeddedConfigYAML())
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("bootstrap config: %w", err)
	}
	result.WroteConfig = wrote

	polPath, err := PolicyPath()
	if err != nil {
		return BootstrapResult{}, err
	}
	wrote, err = writeIfMissing(ctx, polPath, EmbeddedPolicyYAML())
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("bootstrap policy: %w", err)
	}
	result.WrotePolicy = wrote

	// system_profile.json is created by clx doctor or LoadOrDetect on first pipeline run.

	return result, nil
}

func writeIfMissing(ctx context.Context, path string, data []byte) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return false, err
	}

	tmp, err := os.CreateTemp(dir, ".clx-bootstrap-*")
	if err != nil {
		return false, err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return false, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return false, err
	}
	if err := os.Chmod(tmpPath, filePerm); err != nil {
		cleanup()
		return false, err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return false, err
	}
	return true, nil
}
