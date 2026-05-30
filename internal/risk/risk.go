package risk

// RiskLevel classifies command danger.
type RiskLevel int

const (
	Low RiskLevel = iota
	Medium
	High
)

// String returns the level name.
func (l RiskLevel) String() string {
	switch l {
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	default:
		return "unknown"
	}
}

// RiskAssessment is the output of risk classification.
// Confirmation is enforced by config.DecideSafetyAction, not this struct.
type RiskAssessment struct {
	Level  RiskLevel
	Reason string
}
