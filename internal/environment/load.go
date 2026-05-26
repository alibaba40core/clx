package environment

import (
	"context"
	"errors"
	"os"

	"github.com/alibaba40core/clx/internal/config"
)

// LoadOrDetect loads the persisted profile for the current OS/shell, or detects and saves one.
func LoadOrDetect(ctx context.Context) (SystemProfile, error) {
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
	switch {
	case errors.Is(err, os.ErrNotExist):
		store = NewProfileStore()
	case err != nil:
		return SystemProfile{}, err
	}

	if p, ok := store.Profiles[key]; ok && p.OS == curOS && p.Shell == curShell {
		return p, nil
	}

	p, err := Detect(ctx)
	if err != nil {
		return SystemProfile{}, err
	}

	store.UpsertProfile(p)
	if saveErr := SaveStore(ctx, path, store); saveErr != nil {
		// Return detected profile even when persistence fails.
		return p, nil
	}
	return p, nil
}
