package policy

import "errors"

// ErrBlocked is returned when policy denies a command.
var ErrBlocked = errors.New("blocked by policy")

// Result is the outcome of a policy check.
type Result struct {
	Allowed bool
	Reason  string
}

// AllowedResult returns an allow result.
func AllowedResult() Result {
	return Result{Allowed: true}
}
