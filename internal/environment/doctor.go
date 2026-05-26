package environment

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alibaba40core/clx/internal/config"
)

// DoctorOptions configures RunDoctor behavior.
type DoctorOptions struct {
	// Refresh forces a full re-detect for the current shell (sibling shell entries are preserved).
	Refresh bool
}

// RunDoctor detects the environment, saves the profile store, and prints a summary.
func RunDoctor(ctx context.Context, w io.Writer, opts DoctorOptions) error {
	_ = opts.Refresh // explicit refresh always re-detects; flag documents user intent.

	profile, err := Detect(ctx)
	if err != nil {
		return err
	}

	path, err := config.SystemProfilePath()
	if err != nil {
		return err
	}

	store, err := LoadStore(ctx, path)
	switch {
	case os.IsNotExist(err):
		store = NewProfileStore()
	case err != nil:
		return err
	}

	store.UpsertProfile(profile)
	if err := SaveStore(ctx, path, store); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}

	return printSummary(w, profile, path)
}

func printSummary(w io.Writer, p SystemProfile, path string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "CLX environment profile written to %s\n\n", path)
	fmt.Fprintf(&b, "  OS:       %s %s\n", p.OS, p.OSVersion)
	fmt.Fprintf(&b, "  Shell:    %s", p.Shell)
	if p.ShellVersion != "" {
		fmt.Fprintf(&b, " %s", p.ShellVersion)
	}
	b.WriteByte('\n')
	fmt.Fprintf(&b, "  Terminal: %s\n", p.Terminal)
	if p.WSLEnabled {
		b.WriteString("  WSL:      enabled\n")
	}
	if len(p.PackageManagers) > 0 {
		fmt.Fprintf(&b, "  Package managers: %s\n", strings.Join(p.PackageManagers, ", "))
	}
	fmt.Fprintf(&b, "  Tools found: %d (%s)\n", len(p.AvailableTools), strings.Join(p.AvailableTools, ", "))
	if home := p.Paths["home"]; home != "" {
		fmt.Fprintf(&b, "  Home:     %s\n", home)
	}
	if ws := p.Paths["workspace"]; ws != "" {
		fmt.Fprintf(&b, "  Workspace: %s\n", ws)
	}
	_, err := io.WriteString(w, b.String())
	return err
}
