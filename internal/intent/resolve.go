package intent

// IntentSource identifies how an intent was resolved.
type IntentSource int

const (
	SourceRule IntentSource = iota
)

// String returns the source name.
func (s IntentSource) String() string {
	switch s {
	case SourceRule:
		return "Rule"
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
