package aliases

import (
	"strings"

	"github.com/alibaba40core/clx/internal/intent"
	"github.com/alibaba40core/clx/internal/parser"
)

// CollisionWarnings returns reasons the alias name may shadow built-in behavior.
func CollisionWarnings(name string, eng *intent.Engine) []string {
	name = normalizeName(name)
	if name == "" {
		return nil
	}
	var out []string
	if parser.IsKnownShellVerb(name) {
		out = append(out, "name matches a known shell verb")
	}
	if eng == nil {
		return out
	}
	seen := make(map[string]struct{})
	for _, rule := range eng.Rules() {
		for _, ex := range rule.Examples {
			head := exampleHead(ex)
			if head == "" {
				continue
			}
			if _, ok := seen[head]; ok {
				continue
			}
			seen[head] = struct{}{}
			if head == name {
				out = append(out, "name matches rule example head for intent "+rule.Intent)
			}
		}
	}
	return out
}

func exampleHead(example string) string {
	example = strings.TrimSpace(example)
	if example == "" {
		return ""
	}
	if i := strings.IndexByte(example, ' '); i >= 0 {
		return strings.ToLower(example[:i])
	}
	return strings.ToLower(example)
}
