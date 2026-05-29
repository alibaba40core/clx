package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/generator"
)

const explainSchemaVersion = 1

// ExplainEntry is one cached AI explanation for a translated command.
type ExplainEntry struct {
	Key       string    `json:"key"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

type explainFileSchema struct {
	SchemaVersion int            `json:"schema_version"`
	Entries       []ExplainEntry `json:"entries"`
}

// ExplainStore is a bounded LRU explanation cache backed by explanations.json.
type ExplainStore struct {
	path   string
	cfg    config.CacheConfig
	logger *slog.Logger
	now    func() time.Time

	mu     sync.Mutex
	loaded bool
	data   explainFileSchema
}

// LoadExplain opens or creates an explanation store at path.
func LoadExplain(ctx context.Context, path string, cfg config.CacheConfig, logger *slog.Logger) (*ExplainStore, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &ExplainStore{
		path:   path,
		cfg:    cfg,
		logger: logger,
		now:    time.Now,
		data: explainFileSchema{
			SchemaVersion: explainSchemaVersion,
			Entries:       make([]ExplainEntry, 0, 32),
		},
	}, nil
}

// ExplainKeyFor derives the cache key from intent name and generated command.
func ExplainKeyFor(intent string, gen generator.GeneratedCommand) string {
	shell := gen.Shell
	var b strings.Builder
	b.WriteString(intent)
	b.WriteByte(0)
	b.WriteString(shell)
	b.WriteByte(0)
	b.WriteString(strings.Join(gen.Argv, "\x00"))
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// Lookup returns a non-expired entry and bumps LastUsed in memory only.
func (s *ExplainStore) Lookup(ctx context.Context, key string) (ExplainEntry, bool) {
	if err := ctx.Err(); err != nil {
		return ExplainEntry{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedLocked(ctx); err != nil {
		return ExplainEntry{}, false
	}

	now := s.now()
	ttl := time.Duration(s.cfg.TTLDays) * 24 * time.Hour

	for i, e := range s.data.Entries {
		if e.Key != key {
			continue
		}
		if ttl > 0 && now.Sub(e.CreatedAt) > ttl {
			s.data.Entries = append(s.data.Entries[:i], s.data.Entries[i+1:]...)
			return ExplainEntry{}, false
		}
		e.LastUsed = now
		s.data.Entries[i] = e
		s.promoteLocked(i)
		return e, true
	}
	return ExplainEntry{}, false
}

// Put inserts or updates an entry and persists atomically.
func (s *ExplainStore) Put(ctx context.Context, key, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedLocked(ctx); err != nil {
		return err
	}

	now := s.now()
	found := false
	for i, e := range s.data.Entries {
		if e.Key != key {
			continue
		}
		e.Text = text
		e.LastUsed = now
		if e.CreatedAt.IsZero() {
			e.CreatedAt = now
		}
		s.data.Entries[i] = e
		s.promoteLocked(i)
		found = true
		break
	}
	if !found {
		s.data.Entries = append([]ExplainEntry{{
			Key:       key,
			Text:      text,
			CreatedAt: now,
			LastUsed:  now,
		}}, s.data.Entries...)
	}

	s.evictLocked()
	return s.saveLocked(ctx)
}

func (s *ExplainStore) ensureLoadedLocked(ctx context.Context) error {
	if s.loaded {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.loaded = true
	s.data = explainFileSchema{
		SchemaVersion: explainSchemaVersion,
		Entries:       make([]ExplainEntry, 0, 32),
	}

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		s.logWarn("explain cache open failed, starting empty", "err", err)
		return nil
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, maxReadBytes))
	if err != nil {
		s.logWarn("explain cache read failed, starting empty", "err", err)
		return nil
	}

	var loaded explainFileSchema
	if err := json.Unmarshal(raw, &loaded); err != nil {
		s.logWarn("explain cache corrupt, starting empty", "err", err)
		return nil
	}
	if loaded.SchemaVersion != explainSchemaVersion {
		s.logWarn("explain cache schema mismatch, starting empty", "got", loaded.SchemaVersion)
		return nil
	}
	if loaded.Entries == nil {
		loaded.Entries = make([]ExplainEntry, 0)
	}
	s.data = loaded
	return nil
}

func (s *ExplainStore) evictLocked() {
	now := s.now()
	ttl := time.Duration(s.cfg.TTLDays) * 24 * time.Hour

	kept := s.data.Entries[:0]
	for _, e := range s.data.Entries {
		if ttl > 0 && now.Sub(e.CreatedAt) > ttl {
			continue
		}
		kept = append(kept, e)
	}
	s.data.Entries = kept

	for len(s.data.Entries) > s.cfg.MaxEntries {
		s.dropOldestExplainLocked()
	}
	for s.explainEncodedSizeLocked() > int64(s.cfg.MaxDiskBytes) && len(s.data.Entries) > 0 {
		s.dropOldestExplainLocked()
	}
}

func (s *ExplainStore) dropOldestExplainLocked() {
	if len(s.data.Entries) == 0 {
		return
	}
	oldest := 0
	for i := 1; i < len(s.data.Entries); i++ {
		if s.data.Entries[i].LastUsed.Before(s.data.Entries[oldest].LastUsed) {
			oldest = i
		}
	}
	s.data.Entries = append(s.data.Entries[:oldest], s.data.Entries[oldest+1:]...)
}

func (s *ExplainStore) explainEncodedSizeLocked() int64 {
	data, err := json.Marshal(s.data)
	if err != nil {
		return int64(len(s.data.Entries)) * 256
	}
	return int64(len(data))
}

func (s *ExplainStore) promoteLocked(index int) {
	if index <= 0 || index >= len(s.data.Entries) {
		return
	}
	e := s.data.Entries[index]
	copy(s.data.Entries[1:index+1], s.data.Entries[0:index])
	s.data.Entries[0] = e
}

func (s *ExplainStore) saveLocked(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.data.SchemaVersion = explainSchemaVersion
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("explain cache encode: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		s.logWarn("explain cache mkdir failed", "err", err)
		return err
	}

	tmp, err := os.CreateTemp(dir, ".clx-explain-*")
	if err != nil {
		s.logWarn("explain cache temp file failed", "err", err)
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		s.logWarn("explain cache write failed", "err", err)
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		s.logWarn("explain cache close failed", "err", err)
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		cleanup()
		s.logWarn("explain cache chmod failed", "err", err)
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		if rmErr := os.Remove(s.path); rmErr != nil && !os.IsNotExist(rmErr) {
			cleanup()
			s.logWarn("explain cache rename failed", "err", err)
			return err
		}
		if err := os.Rename(tmpPath, s.path); err != nil {
			cleanup()
			s.logWarn("explain cache rename failed", "err", err)
			return err
		}
	}
	return nil
}

func (s *ExplainStore) logWarn(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}
