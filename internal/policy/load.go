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

const policyMissingModNS int64 = -1

// File holds parsed policy rules.
type File struct {
	Blocked []string
	Allowed []string
}

var (
	cacheMu     sync.Mutex
	cached      File
	cachedModNS int64
	loadErr     error
)

// Load reads the policy file from ~/.clx/policies/policy.yaml, reloading when mtime changes.
func Load(ctx context.Context) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}

	path, err := config.PolicyPath()
	if err != nil {
		return File{}, err
	}

	modNS, err := policyModTime(path)
	if err != nil {
		return File{}, err
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()
	if modNS == cachedModNS {
		return cached, loadErr
	}

	if modNS == policyMissingModNS {
		cached = File{}
		loadErr = nil
		cachedModNS = policyMissingModNS
		return cached, nil
	}

	cached, loadErr = loadFile(ctx, path)
	cachedModNS = modNS
	return cached, loadErr
}

func policyModTime(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return policyMissingModNS, nil
		}
		return 0, err
	}
	return fi.ModTime().UnixNano(), nil
}

func loadFile(ctx context.Context, path string) (File, error) {
	if err := ctx.Err(); err != nil {
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
	cacheMu.Lock()
	cached = File{}
	cachedModNS = 0
	loadErr = nil
	cacheMu.Unlock()
}
