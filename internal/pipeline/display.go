package pipeline

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/risk"
)

func printDisplay(w io.Writer, req parser.Request, resolved intent.ResolvedIntent, gen generator.GeneratedCommand, profile environment.SystemProfile, ra risk.RiskAssessment) error {
	quoted := formatCommandForDisplay(gen, profile)

	var b strings.Builder
	if req.EffectiveInput != "" && req.EffectiveInput != strings.TrimSpace(req.RawInput) {
		fmt.Fprintf(&b, "Expanded:    %s\n", req.EffectiveInput)
	}
	fmt.Fprintf(&b, "Intent:      %s\n", resolved.Intent)
	if len(resolved.Params) > 0 {
		keys := make([]string, 0, len(resolved.Params))
		for k := range resolved.Params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Fprint(&b, "Params:      ")
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%s=%q", k, resolved.Params[k])
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "Source:      %s\n", resolved.Source)
	fmt.Fprintf(&b, "Command:     %s\n", quoted)
	fmt.Fprintf(&b, "Explanation: %s\n", gen.Explanation)
	fmt.Fprintf(&b, "Risk:        %s (%s)\n", ra.Level, ra.Reason)
	_, err := io.WriteString(w, b.String())
	return err
}

// formatCommandForDisplay shows argv-quoted or full chain invocation (same as dry-run).
func formatCommandForDisplay(gen generator.GeneratedCommand, profile environment.SystemProfile) string {
	if gen.Chain != nil {
		if inv, err := executor.FormatInvocation(gen, profile); err == nil && inv != "" {
			return inv
		}
	}
	if len(gen.Argv) > 0 {
		return executor.QuoteArgv(gen.Shell, gen.Argv)
	}
	return gen.Command
}
