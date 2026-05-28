package generator

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/capabilities"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/tokenize"
)

// Render builds a GeneratedCommand from a resolved intent and selected strategy.
func Render(ctx context.Context, resolved intent.ResolvedIntent, selected capabilities.SelectedStrategy, profile environment.SystemProfile) (GeneratedCommand, error) {
	if err := ctx.Err(); err != nil {
		return GeneratedCommand{}, err
	}

	params := effectiveParams(resolved.Intent, resolved.Params)
	var argv []string

	if len(selected.Strategy.Argv) > 0 {
		argv = make([]string, 0, len(selected.Strategy.Argv))
		for _, slot := range selected.Strategy.Argv {
			if err := ctx.Err(); err != nil {
				return GeneratedCommand{}, err
			}
			sub, err := substituteSlot(slot, params)
			if err != nil {
				return GeneratedCommand{}, err
			}
			argv = append(argv, sub)
		}
	} else {
		primary, err := substituteSlot(selected.Strategy.Primary, params)
		if err != nil {
			return GeneratedCommand{}, err
		}
		argv, err = tokenize.Tokenize(primary)
		if err != nil {
			return GeneratedCommand{}, err
		}
	}

	return GeneratedCommand{
		Argv:        argv,
		Command:     strings.Join(argv, " "),
		Shell:       profile.Shell,
		Explanation: explanationFor(resolved.Intent),
		ExecHost:    inferExecHost(selected.Key),
	}, nil
}
