package intent

// Rule is a single intent definition from rules/*.yaml or skills/*/intents.yaml.
type Rule struct {
	Intent     string
	SkillPack  string // set for rules loaded from skills/<pack>/intents.yaml
	Examples   []string
	Params     []string
	Strategies map[string]Strategy
}

// Strategy holds a command template for a target OS/shell key.
type Strategy struct {
	Primary      string
	Argv         []string
	RequiresTool string
	Priority     int
}
