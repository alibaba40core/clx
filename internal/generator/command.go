package generator

// GeneratedCommand is a rendered native command ready for display or argv execution.
type GeneratedCommand struct {
	Argv        []string
	Command     string
	Shell       string
	Explanation string
	ExecHost    ExecHost
	Intent      string // optional; set by pipeline for AI explain only
	// AIGenerated marks a command produced directly by an AI provider (no rule
	// backing). Such commands are untrusted: callers must have validated the
	// argv and must keep risk/policy/confirm gating before exec.
	AIGenerated bool
}
