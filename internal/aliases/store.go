package aliases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/alibaba40core/clx/internal/config"
)

// Store is a lazily loaded alias table backed by ~/.clx/aliases.yaml.
type Store struct {
	path       string
	maxAliases int

	mu   sync.Mutex
	file File
}

// Open loads or creates a store at the default aliases path.
func Open(ctx context.Context, maxAliases int) (*Store, error) {
	path, err := config.AliasesPath()
	if err != nil {
		return nil, err
	}
	return OpenAt(ctx, path, maxAliases)
}

// OpenAt opens a store at path.
func OpenAt(ctx context.Context, path string, maxAliases int) (*Store, error) {
	if maxAliases <= 0 {
		maxAliases = 256
	}
	s := &Store{path: path, maxAliases: maxAliases}
	if err := s.reload(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

// Lookup returns the expansion value for a normalized alias name.
func (s *Store) Lookup(name string) (string, bool) {
	name = normalizeName(name)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.file.Entries {
		if e.Name == name {
			return e.Value, true
		}
	}
	return "", false
}

// List returns a copy of all aliases sorted by name.
func (s *Store) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, len(s.file.Entries))
	copy(out, s.file.Entries)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Set adds or updates an alias.
func (s *Store) Set(ctx context.Context, name, value string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := validateValue(value); err != nil {
		return err
	}
	name = normalizeName(name)
	value = strings.TrimSpace(value)

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.file.Entries {
		if e.Name == name {
			s.file.Entries[i].Value = value
			return s.saveLocked(ctx)
		}
	}
	if len(s.file.Entries) >= s.maxAliases {
		return ErrLimitExceeded
	}
	s.file.Entries = append(s.file.Entries, Entry{Name: name, Value: value})
	sort.Slice(s.file.Entries, func(i, j int) bool {
		return s.file.Entries[i].Name < s.file.Entries[j].Name
	})
	return s.saveLocked(ctx)
}

// Remove deletes an alias by name.
func (s *Store) Remove(ctx context.Context, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	name = normalizeName(name)

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.file.Entries {
		if e.Name == name {
			s.file.Entries = append(s.file.Entries[:i], s.file.Entries[i+1:]...)
			return s.saveLocked(ctx)
		}
	}
	return ErrNotFound
}

func (s *Store) reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := readFile(ctx, s.path)
	if err != nil {
		return err
	}
	if f.SchemaVersion != 0 && f.SchemaVersion != SchemaVersion {
		return fmt.Errorf("aliases: unsupported schema_version %d", f.SchemaVersion)
	}
	s.file = f
	s.file.SchemaVersion = SchemaVersion
	return nil
}

func (s *Store) saveLocked(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.file.SchemaVersion = SchemaVersion

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".clx-aliases-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if err := encodeFile(s.file, tmp); err != nil {
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
	if err := os.Rename(tmpPath, s.path); err != nil {
		cleanup()
		return err
	}
	return nil
}
