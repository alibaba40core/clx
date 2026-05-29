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
	resolved, err := resolveChain(ctx, req, resolvers, opts.Logger)
	if err != nil {
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

	// effectiveDryRun is true if either the --dry-run flag or safety.dry_run
	// in config is set. Flag-on or config-on triggers dry-run; both off
	// proceeds to confirm/exec. -y does NOT bypass a config-driven dry-run.
	// TODO(phase3): also derive from cfg.Safety.Mode (medium/high imply
	// dry-run by default). See SafetyConfig godoc for the matrix.
	effectiveDryRun := opts.DryRun || cfg.Safety.DryRun

	if cfg.Features.Explain || opts.Explain || effectiveDryRun || !opts.Yes {
		displayGen := gen
		if shouldEnrichExplanation(opts, resolved) {
			displayGen.Explanation = enrichExplanation(ctx, opts, resolved, gen)
		}
		if err := printDisplay(opts.Stdout, resolved, displayGen, ra); err != nil {
			return 1, err
		}
	}

	if opts.Explain {
		return 0, nil
	}

	if effectiveDryRun {
		inv, err := executor.FormatInvocation(gen, profile)
		if err != nil {
			inv = executor.QuoteArgv(gen.Shell, gen.Argv)
		}
		fmt.Fprintf(opts.Stdout, "dry-run: would execute: %s\n", inv)
		return 0, nil
	}

	needConfirm := !opts.Yes && !cfg.Execution.AutoExecute &&
		(ra.RequiresConfirmation || cfg.Safety.RequireConfirmation)

	if needConfirm {
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
