package environment

import (
	"context"
	"sort"
)

// Detect builds a SystemProfile for the current machine.
func Detect(ctx context.Context) (SystemProfile, error) {
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	paths, err := detectPaths()
	if err != nil {
		return SystemProfile{}, err
	}
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	pkgMgrs := detectPackageManagers()
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	tools := detectTools()
	if err := ctx.Err(); err != nil {
		return SystemProfile{}, err
	}

	return SystemProfile{
		SchemaVersion:   SchemaVersion,
		OS:              detectOS(),
		OSVersion:       detectOSVersion(),
		Shell:           detectShell(),
		ShellVersion:    detectShellVersion(),
		Terminal:        detectTerminal(),
		PackageManagers: sortUnique(pkgMgrs),
		AvailableTools:  sortUnique(tools),
		WSLEnabled:      detectWSL(),
		Paths:           paths,
	}, nil
}

func sortUnique(in []string) []string {
	if len(in) == 0 {
		return in
	}
	sort.Strings(in)
	out := make([]string, 0, len(in))
	var prev string
	for _, s := range in {
		if s == prev {
			continue
		}
		out = append(out, s)
		prev = s
	}
	return out
}
