package intent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/alibaba40core/clx/internal/builtin"
	"github.com/alibaba40core/clx/internal/parser"
)

// ErrNotFound is returned when no rule matches the request.
var ErrNotFound = errors.New("intent not found")

// Engine resolves intents using loaded rules.
type Engine struct {
	rules    []Rule
	byIntent map[string]Rule
	index    patternIndex
}

// Rules returns a snapshot of loaded rules (read-only use).
func (e *Engine) Rules() []Rule {
	if e == nil {
		return nil
	}
	out := make([]Rule, len(e.rules))
	copy(out, e.rules)
	return out
}

// NewEngine returns an engine with the given rules (later entries override duplicate intents).
func NewEngine(rules []Rule) *Engine {
	merged := mergeRules(rules)
	return &Engine{
		rules:    merged,
		byIntent: buildIntentMap(merged),
		index:    compileRules(merged),
	}
}

// NewDefaultEngine loads built-in rules and skills embedded in the binary.
// Use this for production CLI paths where cwd may be outside the source tree.
func NewDefaultEngine() (*Engine, error) {
	rules, err := LoadRulesFromFS(builtin.FS, "rules")
	if err != nil {
		return nil, fmt.Errorf("load builtin rules: %w", err)
	}
	skills, err := LoadSkillsFromFS(builtin.FS, "skills")
	if err != nil {
		return nil, fmt.Errorf("load builtin skills: %w", err)
	}
	return NewEngine(append(rules, skills...)), nil
}

// NewEngineWithOverlay loads embedded built-in rules and skills, then overlays
// optional user content from ~/.clx/rules and ~/.clx/skills. Missing overlay
// directories are silent; malformed overlay files are skipped with a warning.
func NewEngineWithOverlay(ctx context.Context, logger *slog.Logger) (*Engine, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rules, err := loadBuiltinRulesAndSkills()
	if err != nil {
		return nil, err
	}
	rules = appendUserOverlay(ctx, logger, rules)
	return NewEngine(rules), nil
}

// NewEngineFromModuleRoot loads built-in rules and skills from the module root
// (go.mod directory). Intended for tests and rule iteration without rebuilding
// the binary; production uses NewEngineWithOverlay.
func NewEngineFromModuleRoot() (*Engine, error) {
	root, err := findModuleRoot()
	if err != nil {
		return nil, err
	}
	fsys := os.DirFS(root)
	ruleList, err := LoadRulesFromFS(fsys, "internal/builtin/rules")
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}
	skillRules, err := LoadSkillsFromFS(fsys, "internal/builtin/skills")
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

// KnownIntents returns the set of intent names the engine resolves.
// AI providers use this as their closed vocabulary, per
// .cursor/rules/safe-command-execution.mdc.
func (e *Engine) KnownIntents() []string {
	names := make([]string, 0, len(e.byIntent))
	for name := range e.byIntent {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ValidateResolved checks an externally-sourced ResolvedIntent against the
// rule schema: intent name must exist; params must match declared schema.
// Call for every non-Rule source before generator.Translate.
func (e *Engine) ValidateResolved(r ResolvedIntent) error {
	rule, ok := e.RuleForIntent(r.Intent)
	if !ok {
		return fmt.Errorf("unknown intent %q", r.Intent)
	}
	return validateResolvedParams(rule, r.Params)
}

// SkillPacks returns sorted unique skill pack names from loaded rules.
func (e *Engine) SkillPacks() []string {
	if e == nil {
		return nil
	}
	seen := make(map[string]struct{})
	for _, r := range e.rules {
		if r.SkillPack == "" {
			continue
		}
		seen[r.SkillPack] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RuleForIntent returns the rule definition for an intent name.
func (e *Engine) RuleForIntent(name string) (Rule, bool) {
	if e == nil || e.byIntent == nil {
		return Rule{}, false
	}
	r, ok := e.byIntent[name]
	return r, ok
}

// Resolve matches a parsed request to a rule-backed intent.
func (e *Engine) Resolve(ctx context.Context, req parser.Request) (ResolvedIntent, error) {
	if err := ctx.Err(); err != nil {
		return ResolvedIntent{}, err
	}
	if len(req.Tokens) == 0 {
		return ResolvedIntent{}, ErrNotFound
	}

	if hit, ok := e.index.exact[tokenKey(req.Tokens)]; ok {
		return ResolvedIntent{
			Intent:     hit.intent,
			Params:     hit.params,
			Confidence: 1.0,
			Source:     SourceRule,
		}, nil
	}

	for _, comp := range e.index.candidates(req.Tokens) {
		if err := ctx.Err(); err != nil {
			return ResolvedIntent{}, err
		}
		params, ok := matchCompiled(comp, req.Tokens)
		if !ok {
			continue
		}
		if err := validateParams(comp.rule, params); err != nil {
			continue
		}
		return ResolvedIntent{
			Intent:     comp.rule.Intent,
			Params:     params,
			Confidence: 1.0,
			Source:     SourceRule,
		}, nil
	}
	return ResolvedIntent{}, ErrNotFound
}

// RuleNeedsTools reports whether any strategy for intent requires tool probing.
func RuleNeedsTools(eng *Engine, intentName string) bool {
	rule, ok := eng.RuleForIntent(intentName)
	if !ok {
		return false
	}
	for _, strat := range rule.Strategies {
		if strat.RequiresTool != "" {
			return true
		}
	}
	return false
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
