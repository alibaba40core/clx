package intent

// Rule is a single intent definition from rules/*.yaml or skills/*/intents.yaml.
type Rule struct {
	Intent     string
	Examples   []string
	Params     []string
	Strategies map[string]Strategy
}

// Strategy holds a command template for a target OS/shell key (used in Phase 1.5).
type Strategy struct {
	Primary string
}
