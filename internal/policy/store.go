package policy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/alibaba40core/clx/internal/config"
)

const maxAllowedVerbs = 64

var (
	ErrVerbNotFound   = errors.New("allowed verb not found")
	ErrVerbInvalid    = errors.New("invalid verb")
	ErrVerbLimit      = errors.New("allowed verb limit exceeded")
	ErrVerbExists     = errors.New("allowed verb already exists")
)

// DefaultHighAllowedVerbs is the starter allow list for safety.mode=high.
func DefaultHighAllowedVerbs() []string {
	return []string{"git", "docker", "npm"}
}

// EnsureHighDefaults seeds allowed verbs when the list is empty.
func EnsureHighDefaults(ctx context.Context) error {
	pol, err := Load(ctx)
	if err != nil {
		return err
	}
	if len(pol.Allowed) > 0 {
		return nil
	}
	pol.Allowed = append([]string(nil), DefaultHighAllowedVerbs()...)
	return Save(ctx, pol)
}

// ListAllowedVerbs returns sorted allowed verbs from policy file.
func ListAllowedVerbs(ctx context.Context) ([]string, error) {
	pol, err := Load(ctx)
	if err != nil {
		return nil, err
	}
	out := append([]string(nil), pol.Allowed...)
	sort.Strings(out)
	return out, nil
}

// AddAllowedVerb appends a verb to the allow list.
func AddAllowedVerb(ctx context.Context, verb string) error {
	verb, err := normalizeVerb(verb)
	if err != nil {
		return err
	}
	pol, err := Load(ctx)
	if err != nil {
		return err
	}
	for _, v := range pol.Allowed {
		if strings.EqualFold(v, verb) {
			return ErrVerbExists
		}
	}
	if len(pol.Allowed) >= maxAllowedVerbs {
		return ErrVerbLimit
	}
	pol.Allowed = append(pol.Allowed, verb)
	sort.Strings(pol.Allowed)
	return Save(ctx, pol)
}

// RemoveAllowedVerb removes a verb from the allow list.
func RemoveAllowedVerb(ctx context.Context, verb string) error {
	verb, err := normalizeVerb(verb)
	if err != nil {
		return err
	}
	pol, err := Load(ctx)
	if err != nil {
		return err
	}
	idx := -1
	for i, v := range pol.Allowed {
		if strings.EqualFold(v, verb) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrVerbNotFound
	}
	pol.Allowed = append(pol.Allowed[:idx], pol.Allowed[idx+1:]...)
	return Save(ctx, pol)
}

// Save writes policy to ~/.clx/policies/policy.yaml atomically.
func Save(ctx context.Context, pol File) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := config.PolicyPath()
	if err != nil {
		return err
	}
	data := encodePolicyYAML(pol)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".clx-policy-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}
	if _, err := tmp.WriteString(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}
	ResetCache()
	return nil
}

func normalizeVerb(verb string) (string, error) {
	verb = strings.ToLower(strings.TrimSpace(verb))
	if verb == "" || len(verb) > 64 {
		return "", ErrVerbInvalid
	}
	for _, r := range verb {
		if r > unicode.MaxASCII || (!unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_') {
			return "", ErrVerbInvalid
		}
	}
	if strings.ContainsAny(verb, "|;&<>`$") {
		return "", ErrVerbInvalid
	}
	return verb, nil
}

func encodePolicyYAML(pol File) string {
	var b strings.Builder
	b.WriteString("# CLX policy — block list always applies; allowed list only when safety.mode=high\n")
	if pol.AccessLevel != AccessFull {
		fmt.Fprintf(&b, "access_level: %s\n", pol.AccessLevel.String())
	} else {
		b.WriteString("access_level: full\n")
	}
	blocked := pol.Blocked
	if len(blocked) == 0 {
		blocked = defaultBlockedPatterns()
	}
	b.WriteString("\nblocked:\n")
	for _, p := range blocked {
		fmt.Fprintf(&b, "  - %s\n", quoteYAML(p))
	}
	b.WriteString("\n# allowed: enforced only when safety.mode=high (ignored for low/medium/custom)\nallowed:\n")
	for _, v := range pol.Allowed {
		fmt.Fprintf(&b, "  - %s\n", quoteYAML(v))
	}
	return b.String()
}

func defaultBlockedPatterns() []string {
	return []string{`rm -rf /`, "shutdown", "format"}
}

func quoteYAML(s string) string {
	if strings.ContainsAny(s, ":#\n\"' ") || strings.Contains(s, " ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}
