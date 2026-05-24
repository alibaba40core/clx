package cliversion

import "runtime"

// Build metadata — overridden via -ldflags at link time.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = ""
)

// Line returns the human-readable version string for CLX binaries.
func Line(binary string) string {
	built := BuildDate
	if built == "" {
		built = "unknown"
	}
	return binary + " version " + Version + " (commit " + Commit + ", built " + built + ", go " + runtime.Version() + ")"
}
