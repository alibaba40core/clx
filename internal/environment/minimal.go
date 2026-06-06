package environment

import (
	"context"
	"errors"
	"os"

	"github.com/alibaba40core/clx/internal/config"
)

// ErrProfileNotFound is returned when no persisted profile exists for the
// current OS/shell. Run clx doctor to create one.
var ErrProfileNotFound = errors.New("system profile not found: run clx doctor")

// MinimalProfile returns OS and shell from the environment without disk I/O.
// Sufficient for parsing and rule resolution; strategy selection that depends
// on requires_tool needs a full profile from LoadProfile.
func MinimalProfile() SystemProfile {
	return SystemProfile{
		SchemaVersion: SchemaVersion,
		OS:            detectOS(),
		Shell:         detectShell(),
	}
}

// LoadProfile reads the persisted profile for the current OS/shell.
// It never runs Detect; use RunDoctor to create or refresh profiles.
func LoadProfile(ctx context.Context) (SystemProfile, error) {
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		return SystemProfile{}, err
	}

	curOS := detectOS()
	curShell := detectShell()
	key := ProfileKey(curOS, curShell)

	store, err := LoadStore(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SystemProfile{}, ErrProfileNotFound
		}
		return SystemProfile{}, err
	}

	if p, ok := store.Profiles[key]; ok && p.OS == curOS && p.Shell == curShell {
		return p, nil
	}
	return SystemProfile{}, ErrProfileNotFound
}

// ProfileForResolver returns the persisted profile or a minimal env-only profile
// when doctor has not been run. Used by cache and AI resolvers on the slow path.
func ProfileForResolver(ctx context.Context) (SystemProfile, error) {
	p, err := LoadProfile(ctx)
	if err == nil {
		return p, nil
	}
	if errors.Is(err, ErrProfileNotFound) {
		return MinimalProfile(), nil
	}
	return SystemProfile{}, err
}
