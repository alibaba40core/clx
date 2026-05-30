package config

import (
	"fmt"
	"os"
	"sync"
)

var warnLegacySafetyLevel sync.Once

// warnDeprecatedSafetyLevel prints a one-time stderr notice when safety.level is used.
func warnDeprecatedSafetyLevel() {
	warnLegacySafetyLevel.Do(func() {
		_, _ = fmt.Fprintln(os.Stderr, "clx: safety.level is deprecated; use safety.mode instead")
	})
}

// normalizeLegacyMode maps deprecated safety.level values to safety.mode.
func normalizeLegacyMode(v string) string {
	switch v {
	case "safe":
		return "low"
	case "full":
		return "high"
	default:
		return v
	}
}
