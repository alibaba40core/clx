package policy

import (
	"context"
	"os"

	"github.com/alibaba40core/clx/internal/config"
)

// Snapshot is policy file contents plus how they apply for a given safety mode.
type Snapshot struct {
	Path        string
	FileExists  bool
	AccessLevel AccessLevel
	Blocked     []string
	Allowed     []string
	SafetyMode  string
}

// AllowListActive reports whether only listed argv[0] verbs may run at exec time.
func AllowListActive(safetyMode string, allowed []string) bool {
	return enforceAllowList(safetyMode) && len(allowed) > 0
}

// LoadSnapshot reads policy.yaml and pairs it with safety.mode from config.
func LoadSnapshot(ctx context.Context, safetyMode string) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	path, err := config.PolicyPath()
	if err != nil {
		return Snapshot{}, err
	}
	pol, err := Load(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	exists := false
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			exists = true
		}
	}
	return Snapshot{
		Path:        path,
		FileExists:  exists,
		AccessLevel: pol.AccessLevel,
		Blocked:     append([]string(nil), pol.Blocked...),
		Allowed:     append([]string(nil), pol.Allowed...),
		SafetyMode:  safetyMode,
	}, nil
}
