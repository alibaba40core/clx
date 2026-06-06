package environment

import "context"

// LoadOrDetect loads the persisted profile for the current OS/shell.
// When no profile exists it returns ErrProfileNotFound; run clx doctor to create one.
func LoadOrDetect(ctx context.Context) (SystemProfile, error) {
	return LoadProfile(ctx)
}
