package environment

import "os/exec"

// lookPath is overridable in tests.
var lookPath = exec.LookPath

func findOnPath(names []string) []string {
	found := make([]string, 0, len(names))
	for _, name := range names {
		if _, err := lookPath(name); err == nil {
			found = append(found, name)
		}
	}
	return found
}
