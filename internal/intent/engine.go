package intent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alibaba40core/clx/internal/parser"
)

// ErrNotFound is returned when no rule matches the request.
var ErrNotFound = errors.New("intent not found")

// Engine resolves intents using loaded rules.
type Engine struct {
	rules []Rule
}

// NewEngine returns an engine with the given rules (later entries override duplicate intents).
func NewEngine(rules []Rule) *Engine {
	return &Engine{rules: mergeRules(rules)}
}

// NewEngineFromModuleRoot loads rules/ and skills/ under the module root (go.mod directory).
func NewEngineFromModuleRoot() (*Engine, error) {
	root, err := findModuleRoot()
	if err != nil {
		return nil, err
	}
	fsys := os.DirFS(root)
	ruleList, err := LoadRulesFromFS(fsys, "rules")
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}
	skillRules, err := LoadSkillsFromFS(fsys, "skills")
	if err != nil {
		return nil, fmt.Errorf("load skills: %w", err)
	}
	all := append(ruleList, skillRules...)
	return NewEngine(all), nil
}

func mergeRules(rules []Rule) []Rule {
	byIntent := make(map[string]Rule)
	order := make([]string, 0, len(rules))
	for _, r := range rules {
		if _, exists := byIntent[r.Intent]; !exists {
			order = append(order, r.Intent)
		}
		byIntent[r.Intent] = r
	}
	out := make([]Rule, 0, len(order))
	for _, name := range order {
		out = append(out, byIntent[name])
	}
	return out
}

// Resolve matches a parsed request to a rule-backed intent.
func (e *Engine) Resolve(ctx context.Context, req parser.Request) (ResolvedIntent, error) {
	if err := ctx.Err(); err != nil {
		return ResolvedIntent{}, err
	}
	if len(req.Tokens) == 0 {
		return ResolvedIntent{}, ErrNotFound
	}

	for _, rule := range e.rules {
		if err := ctx.Err(); err != nil {
			return ResolvedIntent{}, err
		}
		for _, ex := range rule.Examples {
			params, ok := matchPattern(ex, req.Tokens)
			if !ok {
				continue
			}
			if err := validateParams(rule, params); err != nil {
				continue
			}
			return ResolvedIntent{
				Intent:     rule.Intent,
				Params:     params,
				Confidence: 1.0,
				Source:     SourceRule,
			}, nil
		}
	}
	return ResolvedIntent{}, ErrNotFound
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}
