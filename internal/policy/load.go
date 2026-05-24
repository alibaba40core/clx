package policy

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/yamlutil"
)

const maxPolicyBytes = 32 * 1024

// File holds parsed policy rules.
type File struct {
	Blocked []string
	Allowed []string // ignored in Phase 1.6
}

var (
	loadOnce sync.Once
	cached   File
	loadErr  error
)

// Load reads and caches the policy file from ~/.clx/policies/policy.yaml.
func Load(ctx context.Context) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}
	loadOnce.Do(func() {
		cached, loadErr = loadFile(ctx)
	})
	return cached, loadErr
}

func loadFile(ctx context.Context) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}
	path, err := config.PolicyPath()
	if err != nil {
		return File{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}
		return File{}, err
	}
	defer f.Close()

	root, err := yamlutil.Decode(io.LimitReader(f, maxPolicyBytes))
	if err != nil {
		return File{}, err
	}
	blocked, _ := root.GetStringList("blocked")
	allowed, _ := root.GetStringList("allowed")
	return File{
		Blocked: normalizePatterns(blocked),
		Allowed: normalizePatterns(allowed),
	}, nil
}

func normalizePatterns(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(strings.Trim(s, `"'`))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ResetCache clears the in-process policy cache (tests only).
func ResetCache() {
	loadOnce = sync.Once{}
	cached = File{}
	loadErr = nil
}
