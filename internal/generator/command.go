package generator

// GeneratedCommand is a rendered native command ready for display or argv execution.
type GeneratedCommand struct {
	Argv        []string
	Command     string
	Shell       string
	Explanation string
}
