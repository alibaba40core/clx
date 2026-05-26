package risk

import (
	"context"
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
)

// lowVerbs is the base allowlist of binaries whose default risk is Low.
// A low verb may still be downgraded to Medium when its subverb is not in
// the per-binary read-only allowlist (see gitReadOnlySubverbs /
// dockerReadOnlySubverbs).
var lowVerbs = map[string]struct{}{
	"ls": {}, "grep": {}, "rg": {}, "pwd": {}, "df": {}, "du": {}, "find": {}, "fd": {},
	"dir": {}, "cat": {}, "head": {}, "tail": {}, "git": {},
	"get-location": {}, "get-childitem": {}, "select-string": {}, "findstr": {},
	"docker": {}, "curl": {}, "wget": {},
	"ping": {}, "ss": {}, "netstat": {}, "traceroute": {}, "tracert": {},
}

// gitReadOnlySubverbs lists git subcommands whose default behavior does not
// mutate working tree, refs, or remotes. `git config` is intentionally
// excluded because it mutates state by default.
var gitReadOnlySubverbs = map[string]struct{}{
	"status": {}, "log": {}, "diff": {}, "branch": {},
	"show": {}, "rev-parse": {}, "remote": {}, "tag": {}, "describe": {},
	"ls-files": {}, "ls-remote": {}, "blame": {}, "shortlog": {},
}

// dockerReadOnlySubverbs lists docker subcommands that only inspect state.
var dockerReadOnlySubverbs = map[string]struct{}{
	"ps": {}, "images": {}, "logs": {}, "inspect": {}, "version": {},
	"info": {}, "port": {}, "top": {}, "stats": {}, "events": {},
	"history": {}, "search": {},
}

var destructive = []string{"rm", "shutdown", "format", "del /f", "remove-item", "rmdir"}

// Assess classifies a generated command (Phase 1.6 heuristic stub).
func Assess(ctx context.Context, gen generator.GeneratedCommand) (RiskAssessment, error) {
	if err := ctx.Err(); err != nil {
		return RiskAssessment{}, err
	}

	joined := strings.ToLower(gen.Command)
	for _, d := range destructive {
		if strings.Contains(joined, d) {
			return RiskAssessment{
				Level:                High,
				Reason:               "destructive command pattern",
				RequiresConfirmation: true,
			}, nil
		}
	}
	if strings.Contains(joined, "-rf") || strings.Contains(joined, "/s /q") {
		return RiskAssessment{
			Level:                High,
			Reason:               "recursive or forced delete pattern",
			RequiresConfirmation: true,
		}, nil
	}

	if len(gen.Argv) > 0 {
		verb := strings.ToLower(gen.Argv[0])
		if _, ok := lowVerbs[verb]; ok {
			if reason, ok := nonReadOnlySubverb(verb, gen.Argv); ok {
				return medium(reason), nil
			}
			return RiskAssessment{
				Level:                Low,
				Reason:               "read-only or safe seed command",
				RequiresConfirmation: false,
			}, nil
		}
	}

	return medium("unknown command verb"), nil
}

// nonReadOnlySubverb returns a reason string when a low-verb command uses a
// subverb outside its read-only allowlist. Returns ok=false (and thus Low
// classification stands) when the verb has no subverb gating, when there is
// no subverb, or when the subverb is on the allowlist.
func nonReadOnlySubverb(verb string, argv []string) (string, bool) {
	if len(argv) < 2 {
		return "", false
	}
	sub := strings.ToLower(argv[1])
	switch verb {
	case "git":
		if _, ok := gitReadOnlySubverbs[sub]; !ok {
			return "non-readonly git subcommand: " + sub, true
		}
	case "docker":
		if _, ok := dockerReadOnlySubverbs[sub]; !ok {
			return "non-readonly docker subcommand: " + sub, true
		}
	}
	return "", false
}

func medium(reason string) RiskAssessment {
	return RiskAssessment{
		Level:                Medium,
		Reason:               reason,
		RequiresConfirmation: true,
	}
}
