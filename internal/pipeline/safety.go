package pipeline

import (
	"fmt"
	"io"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
)

// safetyFlagsFromOptions builds config.SafetyFlagOverrides from pipeline options.
func safetyFlagsFromOptions(opts Options) config.SafetyFlagOverrides {
	return config.SafetyFlagOverrides{
		DryRun: opts.DryRun,
		Yes:    opts.Yes,
	}
}

// printDryRunLine writes the preview invocation to stdout.
func printDryRunLine(w io.Writer, gen generator.GeneratedCommand, profile environment.SystemProfile) error {
	inv := formatCommandForDisplay(gen, profile)
	_, err := fmt.Fprintf(w, "dry-run: would execute: %s\n", inv)
	return err
}

// shouldShowDisplay reports whether the command summary should be printed.
func shouldShowDisplay(action config.SafetyAction, opts Options) bool {
	if opts.Explain || !opts.Yes {
		return true
	}
	return action.ShowExplain || action.Preview || action.Confirm
}

// shouldEnrichForSafety returns true when the safety action requests explain enrichment.
func shouldEnrichForSafety(action config.SafetyAction, opts Options, resolved intent.ResolvedIntent, gen generator.GeneratedCommand) bool {
	if gen.AIGenerated {
		return false
	}
	if !action.ShowExplain && !opts.Explain {
		return false
	}
	return shouldEnrichExplanation(opts, resolved)
}
