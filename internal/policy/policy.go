package policy

import "errors"

// ErrBlocked is returned when policy denies a command.
var ErrBlocked = errors.New("blocked by policy")

// Result is the outcome of a policy check.
type Result struct {
	Allowed     bool   // true = pipeline may complete explain path
	ExecAllowed bool   // true = executor may run
	Reason      string
}

// AllowedResult returns a fully allowed result.
func AllowedResult() Result {
	return Result{Allowed: true, ExecAllowed: true}
}

func denyResult(reason string, explainOnly bool) Result {
	if explainOnly {
		return Result{Allowed: true, ExecAllowed: false, Reason: reason}
	}
	return Result{Allowed: false, ExecAllowed: false, Reason: reason}
}
