package cache

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/parser"
)

func testStore(t *testing.T, path string, cfg config.CacheConfig) *Store {
	t.Helper()
	s, err := Load(context.Background(), path, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC) }
	return s
}

func sampleReq() parser.Request {
	return parser.Request{
		RawInput:  "find all logs",
		InputType: parser.InputNaturalLanguage,
		Tokens:    []string{"find", "all", "logs"},
	}
}

func sampleProfile() environment.SystemProfile {
	return environment.SystemProfile{
		OS:             "linux",
		Shell:          "bash",
		AvailableTools: []string{"git", "rg"},
	}
}

func TestKeyForDeterministicAndProfileSensitive(t *testing.T) {
	t.Parallel()
	req := sampleReq()
	p1 := sampleProfile()
	p2 := sampleProfile()
	p2.Shell = "zsh"

	k1 := KeyFor(req, p1)
	k2 := KeyFor(req, p1)
	k3 := KeyFor(req, p2)
	if k1 != k2 {
		t.Fatalf("non-deterministic keys: %q vs %q", k1, k2)
	}
	if k1 == k3 {
		t.Fatal("profile change must change key")
	}
}

func TestLoadMissingFileStartsEmpty(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	s := testStore(t, path, config.Default().Cache)

	key := KeyFor(sampleReq(), sampleProfile())
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected miss on empty store")
	}
}

func TestLoadCorruptFileDegrades(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "intents.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	s := testStore(t, path, config.Default().Cache)
	key := KeyFor(sampleReq(), sampleProfile())
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected miss after corrupt load")
	}
	if err := s.Put(context.Background(), key, "list_dir", map[string]string{}, 0.9); err != nil {
		t.Fatal(err)
	}
	e, ok := s.Lookup(context.Background(), key)
	if !ok || e.Intent != "list_dir" {
		t.Fatalf("entry = %+v ok=%v", e, ok)
	}
}

func TestPutRoundTripPersists(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	cfg := config.Default().Cache
	s := testStore(t, path, cfg)

	key := KeyFor(sampleReq(), sampleProfile())
	if err := s.Put(context.Background(), key, "find_file", map[string]string{"filename": "*.log"}, 0.8); err != nil {
		t.Fatal(err)
	}

	s2 := testStore(t, path, cfg)
	s2.now = s.now
	e, ok := s2.Lookup(context.Background(), key)
	if !ok || e.Intent != "find_file" || e.Params["filename"] != "*.log" {
		t.Fatalf("entry = %+v ok=%v", e, ok)
	}
}

func TestLookupExpiresByTTL(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	cfg := config.Default().Cache
	cfg.TTLDays = 1
	s := testStore(t, path, cfg)

	key := KeyFor(sampleReq(), sampleProfile())
	created := s.now()
	if err := s.Put(context.Background(), key, "list_dir", nil, 0.9); err != nil {
		t.Fatal(err)
	}

	s.now = func() time.Time { return created.Add(48 * time.Hour) }
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected TTL miss")
	}
}

func TestEvictByMaxEntries(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	cfg := config.Default().Cache
	cfg.MaxEntries = 2
	cfg.MaxDiskBytes = 1024 * 1024
	s := testStore(t, path, cfg)

	base := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	step := 0
	s.now = func() time.Time {
		step++
		return base.Add(time.Duration(step) * time.Second)
	}

	keys := []string{"a", "b", "c"}
	for _, k := range keys {
		if err := s.Put(context.Background(), k, "list_dir", nil, 0.9); err != nil {
			t.Fatal(err)
		}
	}

	if _, ok := s.Lookup(context.Background(), "a"); ok {
		t.Fatal("oldest entry should be evicted")
	}
	for _, k := range []string{"b", "c"} {
		if _, ok := s.Lookup(context.Background(), k); !ok {
			t.Fatalf("expected hit for %q", k)
		}
	}
}

func TestEvictByDiskSize(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	cfg := config.Default().Cache
	cfg.MaxEntries = 100
	cfg.MaxDiskBytes = 200
	s := testStore(t, path, cfg)

	for i := 0; i < 5; i++ {
		key := KeyFor(parser.Request{
			InputType: parser.InputNaturalLanguage,
			Tokens:    []string{"phrase", strings.Repeat("x", 40), string(rune('a' + i))},
		}, sampleProfile())
		if err := s.Put(context.Background(), key, "list_dir", map[string]string{"path": strings.Repeat("y", 40)}, 0.9); err != nil {
			t.Fatal(err)
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) > cfg.MaxDiskBytes {
		t.Fatalf("file size %d exceeds cap %d", len(raw), cfg.MaxDiskBytes)
	}
}

func TestSchemaVersionMismatchStartsEmpty(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":99,"entries":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s := testStore(t, path, config.Default().Cache)
	key := KeyFor(sampleReq(), sampleProfile())
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("expected empty store on schema mismatch")
	}
}

func TestPutSkipsSecretShapedParams(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	s := testStore(t, path, config.Default().Cache)

	key := KeyFor(sampleReq(), sampleProfile())
	secretParams := map[string]string{"token": "sk-abcdefghijklmnopqrst"}
	if err := s.Put(context.Background(), key, "find_file", secretParams, 0.9); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Lookup(context.Background(), key); ok {
		t.Fatal("secret-shaped params must not be cached")
	}
}

func TestPutNormalParamsStillCaches(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "intents.json")
	s := testStore(t, path, config.Default().Cache)

	key := KeyFor(sampleReq(), sampleProfile())
	params := map[string]string{"filename": "*.log", "path": "/var/log"}
	if err := s.Put(context.Background(), key, "find_file", params, 0.9); err != nil {
		t.Fatal(err)
	}
	e, ok := s.Lookup(context.Background(), key)
	if !ok || e.Intent != "find_file" || e.Params["filename"] != "*.log" {
		t.Fatalf("entry = %+v ok=%v", e, ok)
	}
}

func TestClearAndAllStats(t *testing.T) {
	t.Setenv("CLX_HOME", t.TempDir())
	ctx := context.Background()
	cfg := config.Default().Cache

	intentsPath, err := config.CacheIntentsPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(intentsPath), 0o700); err != nil {
		t.Fatal(err)
	}
	s := testStore(t, intentsPath, cfg)
	key := KeyFor(sampleReq(), sampleProfile())
	if err := s.Put(ctx, key, "find_file", map[string]string{}, 0.9); err != nil {
		t.Fatal(err)
	}

	stats, err := AllStats(ctx, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 3 || stats[0].Entries != 1 {
		t.Fatalf("stats=%+v", stats)
	}

	if err := ClearAll(ctx, cfg, nil); err != nil {
		t.Fatal(err)
	}
	stats, err = AllStats(ctx, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range stats {
		if st.Entries != 0 {
			t.Fatalf("after clear: %+v", st)
		}
	}
}
