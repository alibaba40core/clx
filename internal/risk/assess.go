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
	"dir": {}, "echo": {}, "cat": {}, "head": {}, "tail": {}, "git": {},
	"get-location": {}, "get-childitem": {}, "select-string": {}, "findstr": {},
	"get-content": {}, "type": {}, "write-output": {},
	"whoami": {}, "date": {}, "env": {}, "which": {}, "where": {}, "uptime": {},
	"get-date": {}, "get-command": {}, "get-uptime": {},
	"docker": {}, "curl": {}, "wget": {},
	"ping": {}, "ss": {}, "netstat": {}, "traceroute": {}, "tracert": {},
	"ipconfig": {}, "ifconfig": {}, "ip": {},
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

// destructiveArgv classifies argv[0] as High regardless of joined command text.
var destructiveArgv = map[string]struct{}{
	"rm": {}, "rmdir": {}, "del": {}, "rd": {}, "remove-item": {},
	"mkfs": {}, "dd": {}, "diskpart": {}, "fdisk": {}, "clear-disk": {},
}

// Assess classifies a generated command from its argv tokens.
func Assess(ctx context.Context, gen generator.GeneratedCommand) (RiskAssessment, error) {
	if err := ctx.Err(); err != nil {
		return RiskAssessment{}, err
	}

	if gen.Chain != nil && len(gen.Chain.Stages) >= 2 {
		return assessChain(ctx, gen), nil
	}

	if len(gen.Argv) > 0 {
		if _, ok := destructiveArgv[strings.ToLower(gen.Argv[0])]; ok {
			return high("destructive command verb"), nil
		}
	}

	if destructiveArgvPattern(gen.Argv) {
		return high("destructive command pattern"), nil
	}
	if recursiveDeletePattern(gen.Argv) {
		return high("recursive or forced delete pattern"), nil
	}
	if removeItemForcedPattern(gen.Argv) {
		return high("powershell forced delete pattern"), nil
	}

	if len(gen.Argv) > 0 {
		verb := strings.ToLower(gen.Argv[0])
		if _, ok := lowVerbs[verb]; ok {
			if reason, ok := nonReadOnlySubverb(verb, gen.Argv); ok {
				return medium(reason), nil
			}
			return RiskAssessment{
				Level:  Low,
				Reason: "read-only or safe seed command",
			}, nil
		}
	}

	return medium("unknown command verb"), nil
}

func high(reason string) RiskAssessment {
	return RiskAssessment{
		Level:  High,
		Reason: reason,
	}
}

// destructiveArgvPattern scans argv tokens (not joined text) for destructive patterns.
func destructiveArgvPattern(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	switch strings.ToLower(argv[0]) {
	case "shutdown", "format", "halt", "reboot", "poweroff":
		return true
	}
	if hasAdjacent(argv, "del", "/f") {
		return true
	}
	if len(argv) >= 2 && strings.EqualFold(argv[0], "docker") && strings.EqualFold(argv[1], "rm") {
		return true
	}
	return false
}

// recursiveDeletePattern detects -rf style flags and /s /q forced deletes in argv.
func recursiveDeletePattern(argv []string) bool {
	if hasFlag(argv, "-rf") || hasFlag(argv, "-fr") {
		return true
	}
	if hasFlag(argv, "-r") && hasFlag(argv, "-f") {
		return true
	}
	if hasAdjacent(argv, "/s", "/q") || hasAdjacent(argv, "/S", "/Q") {
		return true
	}
	if hasAdjacent(argv, "rmdir", "/s") || hasAdjacent(argv, "rmdir", "/S") {
		return true
	}
	return false
}

// removeItemForcedPattern detects Remove-Item with destructive PowerShell flags.
func removeItemForcedPattern(argv []string) bool {
	if len(argv) == 0 || !strings.EqualFold(argv[0], "Remove-Item") {
		return false
	}
	for _, a := range argv[1:] {
		lower := strings.ToLower(a)
		if lower == "-recurse" || lower == "-force" || lower == "-rf" {
			return true
		}
	}
	return false
}

func hasFlag(argv []string, flag string) bool {
	for _, a := range argv {
		if strings.EqualFold(a, flag) {
			return true
		}
	}
	return false
}

func hasAdjacent(argv []string, a, b string) bool {
	for i := 0; i+1 < len(argv); i++ {
		if strings.EqualFold(argv[i], a) && strings.EqualFold(argv[i+1], b) {
			return true
		}
	}
	return false
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
		Level:  Medium,
		Reason: reason,
	}
}
