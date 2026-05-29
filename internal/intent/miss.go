package intent

import "errors"

// MissError is returned when every resolver in the chain returns ErrNotFound.
type MissError struct {
	AIAttempted bool
}

func (e *MissError) Error() string {
	if e.AIAttempted {
		return "ai could not resolve intent"
	}
	return ErrNotFound.Error()
}

func (e *MissError) Is(target error) bool {
	return target == ErrNotFound
}

// AsMiss returns the MissError if err is or wraps one.
func AsMiss(err error) (*MissError, bool) {
	var m *MissError
	if errors.As(err, &m) {
		return m, true
	}
	return nil, false
}
