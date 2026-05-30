package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/executor"
)

const maxSessionFileBytes = 512 * 1024

// Store manages one session file with bounded command history.
type Store struct {
	path      string
	sessionID string
	cfg       config.MemoryConfig
	now       func() time.Time
	mu        sync.Mutex
	sess      Session
}

// Open loads or creates a session store.
func Open(ctx context.Context, sessionID string, cfg config.MemoryConfig) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if sessionID == "" {
		sessionID = DefaultSessionID()
	}
	path, err := SessionPath(sessionID)
	if err != nil {
		return nil, err
	}
	s := &Store{
		path:      path,
		sessionID: sessionID,
		cfg:       cfg,
		now:       time.Now,
	}
	if err := s.load(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

// SessionPath returns ~/.clx/sessions/<id>.json.
func SessionPath(sessionID string) (string, error) {
	dir, err := config.SessionsDir()
	if err != nil {
		return "", err
	}
	safe := sanitizeSessionID(sessionID)
	return filepath.Join(dir, safe+".json"), nil
}

// DefaultSessionID returns the session id for this process.
func DefaultSessionID() string {
	if v := os.Getenv("CLX_SESSION_ID"); v != "" {
		return sanitizeSessionID(v)
	}
	return "default"
}

func sanitizeSessionID(id string) string {
	var b []byte
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b = append(b, byte(r))
		}
	}
	if len(b) == 0 {
		return "default"
	}
	if len(b) > 64 {
		b = b[:64]
	}
	return string(b)
}

func (s *Store) load(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.sess = Empty(s.sessionID)
			return nil
		}
		return err
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, maxSessionFileBytes))
	if err != nil {
		return err
	}
	var loaded Session
	if err := json.Unmarshal(raw, &loaded); err != nil {
		return fmt.Errorf("memory decode: %w", err)
	}
	if loaded.SchemaVersion != 0 && loaded.SchemaVersion != SchemaVersion {
		return fmt.Errorf("memory: unsupported schema_version %d", loaded.SchemaVersion)
	}
	if loaded.Commands == nil {
		loaded.Commands = make([]CommandEntry, 0)
	}
	if loaded.Preferences == nil {
		loaded.Preferences = make(map[string]string)
	}
	s.sess = loaded
	return nil
}

// LastCommand returns the most recent entry, if any.
func (s *Store) LastCommand() (CommandEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sess.Commands) == 0 {
		return CommandEntry{}, false
	}
	return s.sess.Commands[len(s.sess.Commands)-1], true
}

// AppendCommand adds a redacted command entry and persists.
func (s *Store) AppendCommand(ctx context.Context, e CommandEntry) error {
	e.RawInput = executor.Redact(e.RawInput)
	if executor.ContainsSecret(e.RawInput) {
		return nil
	}
	params := make(map[string]string, len(e.Params))
	for k, v := range e.Params {
		if executor.ContainsSecret(v) {
			continue
		}
		params[k] = executor.Redact(v)
	}
	e.Params = params
	e.RecordedAt = s.now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sess.Commands = append(s.sess.Commands, e)
	max := s.cfg.MaxEntriesPerSession
	if max <= 0 {
		max = 64
	}
	if len(s.sess.Commands) > max {
		s.sess.Commands = s.sess.Commands[len(s.sess.Commands)-max:]
	}
	return s.saveLocked(ctx)
}

func (s *Store) saveLocked(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.sess.SchemaVersion = SchemaVersion
	data, err := json.MarshalIndent(s.sess, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".clx-session-*")
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
	return os.Rename(tmpPath, s.path)
}
