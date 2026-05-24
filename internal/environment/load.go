package environment

import (
	"context"
	"os"

	"github.com/alibaba40core/clx/internal/config"
)

// LoadOrDetect loads the persisted system profile, or detects and saves one when missing or empty.
func LoadOrDetect(ctx context.Context) (SystemProfile, error) {
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		return SystemProfile{}, err
	}

	p, err := Load(ctx, path)
	if err != nil {
		if !os.IsNotExist(err) {
			return SystemProfile{}, err
		}
	} else if p.OS != "" {
		return p, nil
	}

	p, err = Detect(ctx)
	if err != nil {
		return SystemProfile{}, err
	}
	if err := Save(ctx, path, p); err != nil {
		return SystemProfile{}, err
	}
	return p, nil
}
