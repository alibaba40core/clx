package config

// normalizeLegacyMode maps deprecated safety.level values to safety.mode.
// Legacy YAML keys are accepted silently until Phase 3 wires Mode to behavior.
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
