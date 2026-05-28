package intent

// IntentSource identifies how an intent was resolved.
type IntentSource int

const (
	SourceRule IntentSource = iota
	SourceCache
	SourceAI
	SourceMemory // Phase 4
)

// String returns the source name.
func (s IntentSource) String() string {
	switch s {
	case SourceRule:
		return "Rule"
	case SourceCache:
		return "Cache"
	case SourceAI:
		return "AI"
	case SourceMemory:
		return "Memory"
	default:
		return "Unknown"
	}
}

// ResolvedIntent is the output of intent resolution.
type ResolvedIntent struct {
	Intent     string
	Params     map[string]string
	Confidence float64
	Source     IntentSource
}
