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
	"strings"
	"sync"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/providers"
)

const commandSchemaVersion = 1

// CommandEntry is one cached AI command-generation result.
type CommandEntry struct {
	Key         string    `json:"key"`
	Argv        []string  `json:"argv"`
	Shell       string    `json:"shell"`
	Explanation string    `json:"explanation"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
}

type commandFileSchema struct {
	SchemaVersion int            `json:"schema_version"`
	Entries       []CommandEntry `json:"entries"`
}

// CommandStore is a bounded LRU command-generation cache backed by commands.json.
type CommandStore struct {
	path   string
	cfg    config.CacheConfig
	logger *slog.Logger
	now    func() time.Time

	mu     sync.Mutex
	loaded bool
	data   commandFileSchema
}

// LoadCommands opens or creates a command cache store at path.
func LoadCommands(ctx context.Context, path string, cfg config.CacheConfig, logger *slog.Logger) (*CommandStore, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &CommandStore{
		path:   path,
		cfg:    cfg,
		logger: logger,
		now:    time.Now,
		data: commandFileSchema{
			SchemaVersion: commandSchemaVersion,
			Entries:       make([]CommandEntry, 0, 32),
		},
	}, nil
}

// CommandKeyFor derives the cache key from raw input and profile grounding.
func CommandKeyFor(raw string, profile environment.SystemProfile) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(raw))
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
func (s *CommandStore) Lookup(ctx context.Context, key string) (CommandEntry, bool) {
	if err := ctx.Err(); err != nil {
		return CommandEntry{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedLocked(ctx); err != nil {
		return CommandEntry{}, false
	}

	now := s.now()
	ttl := time.Duration(s.cfg.TTLDays) * 24 * time.Hour

	for i, e := range s.data.Entries {
		if e.Key != key {
			continue
		}
		if ttl > 0 && now.Sub(e.CreatedAt) > ttl {
			s.data.Entries = append(s.data.Entries[:i], s.data.Entries[i+1:]...)
			return CommandEntry{}, false
		}
		e.LastUsed = now
		s.data.Entries[i] = e
		s.promoteCommandLocked(i)
		return e, true
	}
	return CommandEntry{}, false
}

// Put inserts or updates an entry and persists atomically.
func (s *CommandStore) Put(ctx context.Context, key string, resp *providers.CommandResponse) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if resp == nil {
		return nil
	}
	for _, tok := range resp.Argv {
		if executor.ContainsSecret(tok) {
			s.logDebug("command cache write skipped: secret-shaped argv token")
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedLocked(ctx); err != nil {
		return err
	}

	now := s.now()
	argv := append([]string(nil), resp.Argv...)
	found := false
	for i, e := range s.data.Entries {
		if e.Key != key {
			continue
		}
		e.Argv = argv
		e.Shell = resp.Shell
		e.Explanation = resp.Explanation
		e.Confidence = resp.Confidence
		e.LastUsed = now
		if e.CreatedAt.IsZero() {
			e.CreatedAt = now
		}
		s.data.Entries[i] = e
		s.promoteCommandLocked(i)
		found = true
		break
	}
	if !found {
		s.data.Entries = append([]CommandEntry{{
			Key:         key,
			Argv:        argv,
			Shell:       resp.Shell,
			Explanation: resp.Explanation,
			Confidence:  resp.Confidence,
			CreatedAt:   now,
			LastUsed:    now,
		}}, s.data.Entries...)
	}

	s.evictCommandLocked()
	return s.saveCommandLocked(ctx)
}

// ToCommandResponse builds a provider response from a cache entry.
func ToCommandResponse(e CommandEntry) *providers.CommandResponse {
	argv := append([]string(nil), e.Argv...)
	return &providers.CommandResponse{
		Argv:        argv,
		Shell:       e.Shell,
		Explanation: e.Explanation,
		Confidence:  e.Confidence,
	}
}

func (s *CommandStore) ensureLoadedLocked(ctx context.Context) error {
	if s.loaded {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.loaded = true
	s.data = commandFileSchema{
		SchemaVersion: commandSchemaVersion,
		Entries:       make([]CommandEntry, 0, 32),
	}

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		s.logWarn("command cache open failed, starting empty", "err", err)
		return nil
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, maxReadBytes))
	if err != nil {
		s.logWarn("command cache read failed, starting empty", "err", err)
		return nil
	}

	var loaded commandFileSchema
	if err := json.Unmarshal(raw, &loaded); err != nil {
		s.logWarn("command cache corrupt, starting empty", "err", err)
		return nil
	}
	if loaded.SchemaVersion != commandSchemaVersion {
		s.logWarn("command cache schema mismatch, starting empty", "got", loaded.SchemaVersion)
		return nil
	}
	if loaded.Entries == nil {
		loaded.Entries = make([]CommandEntry, 0)
	}
	s.data = loaded
	return nil
}

func (s *CommandStore) evictCommandLocked() {
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
		s.dropOldestCommandLocked()
	}
	for s.commandEncodedSizeLocked() > int64(s.cfg.MaxDiskBytes) && len(s.data.Entries) > 0 {
		s.dropOldestCommandLocked()
	}
}

func (s *CommandStore) dropOldestCommandLocked() {
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

func (s *CommandStore) commandEncodedSizeLocked() int64 {
	data, err := json.Marshal(s.data)
	if err != nil {
		return int64(len(s.data.Entries)) * 512
	}
	return int64(len(data))
}

func (s *CommandStore) promoteCommandLocked(index int) {
	if index <= 0 || index >= len(s.data.Entries) {
		return
	}
	e := s.data.Entries[index]
	copy(s.data.Entries[1:index+1], s.data.Entries[0:index])
	s.data.Entries[0] = e
}

func (s *CommandStore) saveCommandLocked(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.data.SchemaVersion = commandSchemaVersion
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("command cache encode: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		s.logWarn("command cache mkdir failed", "err", err)
		return err
	}

	tmp, err := os.CreateTemp(dir, ".clx-command-*")
	if err != nil {
		s.logWarn("command cache temp file failed", "err", err)
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		s.logWarn("command cache write failed", "err", err)
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		s.logWarn("command cache close failed", "err", err)
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		cleanup()
		s.logWarn("command cache chmod failed", "err", err)
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		if rmErr := os.Remove(s.path); rmErr != nil && !os.IsNotExist(rmErr) {
			cleanup()
			s.logWarn("command cache rename failed", "err", err)
			return err
		}
		if err := os.Rename(tmpPath, s.path); err != nil {
			cleanup()
			s.logWarn("command cache rename failed", "err", err)
			return err
		}
	}
	return nil
}

func (s *CommandStore) logWarn(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}

func (s *CommandStore) logDebug(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}
