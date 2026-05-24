package generator

import (
	"context"

	"github.com/alibaba40core/clx/internal/capabilities"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

// Translate resolves intent, selects strategy, and renders a command.
func Translate(ctx context.Context, eng *intent.Engine, resolved intent.ResolvedIntent, profile environment.SystemProfile) (GeneratedCommand, error) {
	rule, ok := eng.RuleForIntent(resolved.Intent)
	if !ok {
		return GeneratedCommand{}, intent.ErrNotFound
	}
	selected, err := capabilities.Select(ctx, rule, profile)
	if err != nil {
		return GeneratedCommand{}, err
	}
	return Render(ctx, resolved, selected, profile)
}
