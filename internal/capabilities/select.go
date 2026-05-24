package capabilities

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/intent"
)

// ErrNoStrategy is returned when no strategy matches the profile.
var ErrNoStrategy = errors.New("no strategy for environment")

// Select chooses the best strategy for rule and profile.
func Select(ctx context.Context, rule intent.Rule, profile environment.SystemProfile) (SelectedStrategy, error) {
	if err := ctx.Err(); err != nil {
		return SelectedStrategy{}, err
	}
	if len(rule.Strategies) == 0 {
		return SelectedStrategy{}, ErrNoStrategy
	}

	tools := toolSet(profile.AvailableTools)
	type candidate struct {
		key      string
		strategy intent.Strategy
	}
	var eligible []candidate

	for key, strat := range rule.Strategies {
		if err := ctx.Err(); err != nil {
			return SelectedStrategy{}, err
		}
		if !strategyMatchesKey(key, profile) {
			continue
		}
		if strat.RequiresTool != "" && !tools[strings.ToLower(strat.RequiresTool)] {
			continue
		}
		eligible = append(eligible, candidate{key: key, strategy: strat})
	}

	if len(eligible) == 0 {
		return SelectedStrategy{}, fmt.Errorf("%w: intent %q", ErrNoStrategy, rule.Intent)
	}

	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].strategy.Priority != eligible[j].strategy.Priority {
			return eligible[i].strategy.Priority > eligible[j].strategy.Priority
		}
		ri := keyRank(eligible[i].key, profile)
		rj := keyRank(eligible[j].key, profile)
		if ri != rj {
			return ri < rj
		}
		return eligible[i].key < eligible[j].key
	})

	best := eligible[0]
	return SelectedStrategy{Key: best.key, Strategy: best.strategy}, nil
}

func toolSet(tools []string) map[string]bool {
	m := make(map[string]bool, len(tools))
	for _, t := range tools {
		m[strings.ToLower(t)] = true
	}
	return m
}
