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

// LoadStore reads a profile store from path, migrating v1 flat profiles when needed.
func LoadStore(ctx context.Context, path string) (ProfileStore, error) {
	store, _, err := loadStoreFromPath(ctx, path)
	return store, err
}

func loadStoreFromPath(ctx context.Context, path string) (ProfileStore, bool, error) {
	if err := ctx.Err(); err != nil {
		return ProfileStore{}, false, err
	}

	f, err := os.Open(path)
	if err != nil {
		return ProfileStore{}, false, err
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, maxProfileBytes))
	if err != nil {
		return ProfileStore{}, false, err
	}

	store, migrated, err := migrateV1IfNeeded(raw)
	if err != nil {
		return ProfileStore{}, false, fmt.Errorf("decode profile store: %w", err)
	}
	if migrated {
		if err := saveStore(ctx, path, store); err != nil {
			return store, true, fmt.Errorf("persist migrated profile: %w", err)
		}
	}
	return store, migrated, nil
}

// migrateV1IfNeeded decodes raw JSON as a v2 store or wraps a v1 flat profile.
func migrateV1IfNeeded(raw []byte) (ProfileStore, bool, error) {
	var probe struct {
		Profiles json.RawMessage `json:"profiles"`
		OS       string          `json:"os"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ProfileStore{}, false, err
	}

	if len(probe.Profiles) > 0 && string(probe.Profiles) != "null" {
		var store ProfileStore
		if err := json.Unmarshal(raw, &store); err != nil {
			return ProfileStore{}, false, err
		}
		if store.Profiles == nil {
			store.Profiles = make(map[string]SystemProfile)
		}
		if store.SchemaVersion == 0 {
			store.SchemaVersion = SchemaVersion
		}
		return store, false, nil
	}

	var p SystemProfile
	if err := json.Unmarshal(raw, &p); err != nil {
		return ProfileStore{}, false, err
	}
	if p.OS == "" {
		return ProfileStore{}, false, fmt.Errorf("invalid profile: missing os")
	}

	store := NewProfileStore()
	store.UpsertProfile(p)
	return store, true, nil
}

// SaveStore writes a profile store to path atomically.
func SaveStore(ctx context.Context, path string, store ProfileStore) error {
	return saveStore(ctx, path, store)
}

func saveStore(ctx context.Context, path string, store ProfileStore) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	store.SchemaVersion = SchemaVersion
	if store.Profiles == nil {
		store.Profiles = make(map[string]SystemProfile)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile store: %w", err)
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
