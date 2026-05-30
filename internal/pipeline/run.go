package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/executor"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
	"github.com/alibaba40core/clx/internal/policy"
	"github.com/alibaba40core/clx/internal/risk"
)

// Run executes the full parse → intent → translate → safety → optional exec pipeline.
// Returns exit code (0 success) and an error for fatal failures (message may already be printed).
func Run(ctx context.Context, cfg config.Config, rawInput string, opts Options) (int, error) {
	opts.WithDefaults()

	profile, err := environment.LoadOrDetect(ctx)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "profile: %v\n", err)
		return 1, err
	}

	req, err := parser.Parse(ctx, rawInput, profile)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "parse: %v\n", err)
		return 1, err
	}

	eng := opts.Engine
	if eng == nil {
		var err error
		eng, err = intent.NewEngineWithOverlay(ctx, opts.Logger)
		if err != nil {
			fmt.Fprintf(opts.Stderr, "rules: %v\n", err)
			return 1, err
		}
	}

	resolvers := buildResolvers(eng, opts)
	resolved, err := resolveChain(ctx, req, resolvers, opts.Logger, aiResolverIndex(opts))
	if err != nil {
		// Hybrid fallback: rules/cache/AI-intent all missed. If enabled, ask the
		// provider to generate a full command (argv) for this platform. The argv
		// is validated, risk-assessed, policy-gated, and confirmed before exec.
		if isResolverMiss(err) {
			if code, handled, aiErr := tryAICommand(ctx, cfg, opts, profile, req); handled {
				return code, aiErr
			}
		}
		return reportResolveError(opts, req, err)
	}

	if resolved.Source != intent.SourceRule {
		if err := eng.ValidateResolved(resolved); err != nil {
			if opts.Logger != nil {
				opts.Logger.Warn("non-rule resolver output rejected",
					"source", resolved.Source.String(),
					"intent", resolved.Intent,
					"err", err,
				)
			}
			fmt.Fprintf(opts.Stderr, "untrusted resolver output rejected: %v\n", err)
			return 1, err
		}
	}

	if err := executor.ValidateIntentParams(resolved.Params); err != nil {
		fmt.Fprintf(opts.Stderr, "validation: %v\n", err)
		return 1, err
	}

	gen, err := generator.Translate(ctx, eng, resolved, profile)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "translate: %v\n", err)
		return 1, err
	}

	return executePlan(ctx, cfg, opts, profile, resolved, gen)
}

// isResolverMiss reports whether err means every resolver missed (vs a hard
// provider/transport error), making AI command generation an eligible fallback.
func isResolverMiss(err error) bool {
	if _, ok := intent.AsMiss(err); ok {
		return true
	}
	return errors.Is(err, intent.ErrNotFound)
}

// reportResolveError prints the appropriate message for a resolution failure and
// returns the pipeline exit code.
func reportResolveError(opts Options, req parser.Request, err error) (int, error) {
	if miss, ok := intent.AsMiss(err); ok && miss.AIAttempted {
		if req.InputType == parser.InputNaturalLanguage {
			fmt.Fprintf(opts.Stderr, "AI could not map this to a known command intent; try a simpler phrase or split into separate commands (e.g. clx pwd, then clx \"list directory .\")\n")
		} else {
			fmt.Fprintf(opts.Stderr, "no matching intent after rules and AI\n")
		}
		return 1, err
	}
	if errors.Is(err, intent.ErrNotFound) {
		if req.InputType == parser.InputNaturalLanguage {
			fmt.Fprintf(opts.Stderr, "no matching rule for natural language input; try an explicit command (e.g. grep PATTERN FILE)\n")
		} else {
			fmt.Fprintf(opts.Stderr, "no matching rule for input\n")
		}
		return 1, err
	}
	if msg := err.Error(); strings.Contains(msg, "provider unavailable") {
		fmt.Fprintf(opts.Stderr, "%s\n", msg)
		return 1, err
	}
	if strings.Contains(err.Error(), "provider timeout") {
		fmt.Fprintf(opts.Stderr, "intent: %v\n", err)
		return 1, err
	}
	fmt.Fprintf(opts.Stderr, "intent: %v\n", err)
	return 1, err
}

// executePlan runs the shared safety + execution stage for a generated command,
// whether it came from a rule/intent or from AI command generation:
// risk → policy → dry-run → (risk-based) confirm → argv-only exec.
func executePlan(ctx context.Context, cfg config.Config, opts Options, profile environment.SystemProfile, resolved intent.ResolvedIntent, gen generator.GeneratedCommand) (int, error) {
	ra, err := risk.Assess(ctx, gen)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "risk: %v\n", err)
		return 1, err
	}

	pol, err := policy.Check(ctx, gen, ra)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "policy: %v\n", err)
		return 1, err
	}
	if !pol.Allowed {
		fmt.Fprintf(opts.Stderr, "blocked by policy: %s\n", pol.Reason)
		return 1, policy.ErrBlocked
	}

	flags := safetyFlagsFromOptions(opts)
	action := config.DecideSafetyAction(cfg, ra.Level.String(), flags)

	if shouldShowDisplay(action, opts) {
		displayGen := gen
		if shouldEnrichForSafety(action, opts, resolved, gen) {
			displayGen.Explanation = enrichExplanation(ctx, opts, resolved, gen)
		}
		if err := printDisplay(opts.Stdout, resolved, displayGen, ra); err != nil {
			return 1, err
		}
	}

	if opts.Explain {
		return 0, nil
	}

	if action.Preview || opts.DryRun {
		if err := printDryRunLine(opts.Stdout, gen, profile); err != nil {
			return 1, err
		}
		if action.PreviewOnly(cfg, flags) {
			return 0, nil
		}
	}

	if action.ShouldConfirm(cfg, flags) {
		ok, err := confirmPrompt(opts.Stdin, opts.Stdout)
		if err != nil {
			return 1, err
		}
		if !ok {
			return 0, nil
		}
	}

	timeout := time.Duration(cfg.Execution.Timeout) * time.Second
	if err := executor.Run(ctx, gen,
		executor.WithRisk(ra),
		executor.WithPolicy(pol),
		executor.WithProfile(profile),
		executor.WithTimeout(timeout),
		executor.WithIO(opts.Stdout, opts.Stderr),
	); err != nil {
		fmt.Fprintf(opts.Stderr, "execute: %v\n", err)
		return 1, err
	}
	return 0, nil
}
