package cache

import (
	"context"
	"log/slog"
	"os"

	"github.com/alibaba40core/clx/internal/config"
)

// VolumeStat describes one on-disk cache file.
type VolumeStat struct {
	Name    string
	Path    string
	Entries int
	Bytes   int64
}

// AllStats returns entry counts and file sizes for intent and explanation caches.
func AllStats(ctx context.Context, cfg config.CacheConfig, logger *slog.Logger) ([]VolumeStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	intentsPath, err := config.CacheIntentsPath()
	if err != nil {
		return nil, err
	}
	explainPath, err := config.CacheExplanationsPath()
	if err != nil {
		return nil, err
	}

	intentStore, err := Load(ctx, intentsPath, cfg, logger)
	if err != nil {
		return nil, err
	}
	explainStore, err := LoadExplain(ctx, explainPath, cfg, logger)
	if err != nil {
		return nil, err
	}

	intentEntries, err := intentStore.entryCount(ctx)
	if err != nil {
		return nil, err
	}
	explainEntries, err := explainStore.entryCount(ctx)
	if err != nil {
		return nil, err
	}

	intentBytes, _ := fileSize(intentsPath)
	explainBytes, _ := fileSize(explainPath)

	return []VolumeStat{
		{Name: "intents", Path: intentsPath, Entries: intentEntries, Bytes: intentBytes},
		{Name: "explanations", Path: explainPath, Entries: explainEntries, Bytes: explainBytes},
	}, nil
}

// ClearAll truncates intent and explanation caches on disk.
func ClearAll(ctx context.Context, cfg config.CacheConfig, logger *slog.Logger) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	intentsPath, err := config.CacheIntentsPath()
	if err != nil {
		return err
	}
	explainPath, err := config.CacheExplanationsPath()
	if err != nil {
		return err
	}

	intentStore, err := Load(ctx, intentsPath, cfg, logger)
	if err != nil {
		return err
	}
	if err := intentStore.Clear(ctx); err != nil {
		return err
	}
	explainStore, err := LoadExplain(ctx, explainPath, cfg, logger)
	if err != nil {
		return err
	}
	return explainStore.Clear(ctx)
}

func (s *Store) entryCount(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureLoadedLocked(ctx); err != nil {
		return 0, err
	}
	return len(s.data.Entries), nil
}

// Clear resets the intent cache to an empty schema and persists atomically.
func (s *Store) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	s.loaded = true
	s.data = fileSchema{
		SchemaVersion: SchemaVersion,
		Entries:       make([]Entry, 0),
	}
	return s.saveLocked(ctx)
}

func (s *ExplainStore) entryCount(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureLoadedLocked(ctx); err != nil {
		return 0, err
	}
	return len(s.data.Entries), nil
}

// Clear resets the explanation cache to an empty schema and persists atomically.
func (s *ExplainStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	s.loaded = true
	s.data = explainFileSchema{
		SchemaVersion: explainSchemaVersion,
		Entries:       make([]ExplainEntry, 0),
	}
	return s.saveLocked(ctx)
}

func fileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return fi.Size(), nil
}
