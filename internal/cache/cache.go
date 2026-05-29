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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/parser"
)

const (
	SchemaVersion = 1
	maxReadBytes  = 6 * 1024 * 1024
)

// Entry is one cached intent resolution.
type Entry struct {
	Key        string            `json:"key"`
	Intent     string            `json:"intent"`
	Params     map[string]string `json:"params"`
	Confidence float64           `json:"confidence"`
	CreatedAt  time.Time         `json:"created_at"`
	LastUsed   time.Time         `json:"last_used"`
}

type fileSchema struct {
	SchemaVersion int     `json:"schema_version"`
	Entries       []Entry `json:"entries"`
}

// Store is a bounded LRU intent cache backed by a single JSON file.
type Store struct {
	path   string
	cfg    config.CacheConfig
	logger *slog.Logger
	now    func() time.Time

	mu     sync.Mutex
	loaded bool
	data   fileSchema
}

// Load opens or creates a cache store at path. Missing or corrupt files degrade
// to an empty in-memory store; only ctx cancellation returns an error.
func Load(ctx context.Context, path string, cfg config.CacheConfig, logger *slog.Logger) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s := &Store{
		path:   path,
		cfg:    cfg,
		logger: logger,
		now:    time.Now,
		data: fileSchema{
			SchemaVersion: SchemaVersion,
			Entries:       make([]Entry, 0, 32),
		},
	}
	return s, nil
}

// KeyFor derives the cache key from request tokens and profile grounding.
func KeyFor(req parser.Request, profile environment.SystemProfile) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(int(req.InputType)))
	b.WriteByte(0)
	b.WriteString(strings.Join(req.Tokens, "\x00"))
	b.WriteByte(0)
	b.WriteString(profile.OS)
	b.WriteByte(0)
	b.WriteString(profile.Shell)
	b.WriteByte(0)

	tools := append([]string(nil), profile.AvailableTools...)
	sort.Strings(tools)
	b.WriteString(strings.Join(tools, "\x00"))

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// Lookup returns a non-expired entry and bumps LastUsed in memory only.
func (s *Store) Lookup(ctx context.Context, key string) (Entry, bool) {
	if err := ctx.Err(); err != nil {
		return Entry{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedLocked(ctx); err != nil {
		return Entry{}, false
	}

	now := s.now()
	ttl := time.Duration(s.cfg.TTLDays) * 24 * time.Hour

	for i, e := range s.data.Entries {
		if e.Key != key {
			continue
		}
		if ttl > 0 && now.Sub(e.CreatedAt) > ttl {
			s.data.Entries = append(s.data.Entries[:i], s.data.Entries[i+1:]...)
			return Entry{}, false
		}
		e.LastUsed = now
		s.data.Entries[i] = e
		s.promoteLocked(i)
		return e, true
	}
	return Entry{}, false
}

// Put inserts or updates an entry and persists atomically.
func (s *Store) Put(ctx context.Context, key string, intent string, params map[string]string, confidence float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedLocked(ctx); err != nil {
		return err
	}

	now := s.now()
	if params == nil {
		params = map[string]string{}
	}

	found := false
	for i, e := range s.data.Entries {
		if e.Key != key {
			continue
		}
		e.Intent = intent
		e.Params = cloneParams(params)
		e.Confidence = confidence
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
		s.data.Entries = append([]Entry{{
			Key:        key,
			Intent:     intent,
			Params:     cloneParams(params),
			Confidence: confidence,
			CreatedAt:  now,
			LastUsed:   now,
		}}, s.data.Entries...)
	}

	s.evictLocked()
	return s.saveLocked(ctx)
}

func (s *Store) ensureLoadedLocked(ctx context.Context) error {
	if s.loaded {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.loaded = true
	s.data = fileSchema{
		SchemaVersion: SchemaVersion,
		Entries:       make([]Entry, 0, 32),
	}

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		s.logWarn("cache open failed, starting empty", "err", err)
		return nil
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, maxReadBytes))
	if err != nil {
		s.logWarn("cache read failed, starting empty", "err", err)
		return nil
	}

	var loaded fileSchema
	if err := json.Unmarshal(raw, &loaded); err != nil {
		s.logWarn("cache corrupt, starting empty", "err", err)
		return nil
	}
	if loaded.SchemaVersion != SchemaVersion {
		s.logWarn("cache schema mismatch, starting empty", "got", loaded.SchemaVersion, "want", SchemaVersion)
		return nil
	}
	if loaded.Entries == nil {
		loaded.Entries = make([]Entry, 0)
	}
	s.data = loaded
	return nil
}

func (s *Store) evictLocked() {
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
		s.dropOldestLocked()
	}

	for s.encodedSizeLocked() > int64(s.cfg.MaxDiskBytes) && len(s.data.Entries) > 0 {
		s.dropOldestLocked()
	}
}

func (s *Store) dropOldestLocked() {
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

func (s *Store) encodedSizeLocked() int64 {
	data, err := json.Marshal(s.data)
	if err != nil {
		return int64(len(s.data.Entries)) * 256
	}
	return int64(len(data))
}

func (s *Store) promoteLocked(index int) {
	if index <= 0 || index >= len(s.data.Entries) {
		return
	}
	e := s.data.Entries[index]
	copy(s.data.Entries[1:index+1], s.data.Entries[0:index])
	s.data.Entries[0] = e
}

func (s *Store) saveLocked(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.data.SchemaVersion = SchemaVersion
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("cache encode: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		s.logWarn("cache mkdir failed", "err", err)
		return err
	}

	tmp, err := os.CreateTemp(dir, ".clx-cache-*")
	if err != nil {
		s.logWarn("cache temp file failed", "err", err)
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		s.logWarn("cache write failed", "err", err)
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		s.logWarn("cache close failed", "err", err)
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		cleanup()
		s.logWarn("cache chmod failed", "err", err)
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		if rmErr := os.Remove(s.path); rmErr != nil && !os.IsNotExist(rmErr) {
			cleanup()
			s.logWarn("cache rename failed", "err", err)
			return err
		}
		if err := os.Rename(tmpPath, s.path); err != nil {
			cleanup()
			s.logWarn("cache rename failed", "err", err)
			return err
		}
	}
	return nil
}

func (s *Store) logWarn(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}

func cloneParams(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
