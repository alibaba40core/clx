package policy

import (
	"strings"

	"github.com/alibaba40core/clx/internal/risk"
)

// AccessLevel controls which commands may execute regardless of safety mode UX.
type AccessLevel int

const (
	// AccessFull allows any command that passes block/allow lists (default).
	AccessFull AccessLevel = iota
	// AccessModerate allows only low-risk commands.
	AccessModerate
	// AccessSafe denies all execution (explain-only).
	AccessSafe
)

// String returns the policy.yaml value for the level.
func (l AccessLevel) String() string {
	switch l {
	case AccessSafe:
		return "safe"
	case AccessModerate:
		return "moderate"
	default:
		return "full"
	}
}

func parseAccessLevel(raw string) AccessLevel {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "safe":
		return AccessSafe
	case "moderate":
		return AccessModerate
	default:
		return AccessFull
	}
}

func accessLevelAllows(ra risk.RiskAssessment, level AccessLevel) (bool, string) {
	switch level {
	case AccessSafe:
		return false, "access level safe: explain only"
	case AccessModerate:
		if ra.Level != risk.Low {
			return false, "access level moderate: only low-risk commands allowed"
		}
		return true, ""
	default:
		return true, ""
	}
}
