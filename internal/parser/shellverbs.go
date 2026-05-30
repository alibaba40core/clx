package parser

import "strings"

// shellVerbs is a bounded set of known shell command first tokens.
// Inclusion here prevents the natural-language heuristic from misrouting
// 3+ token commands that legitimately start with these binaries
// (e.g. `ping -c 4 google.com`).
var shellVerbs = map[string]struct{}{
	"git": {}, "docker": {}, "kubectl": {}, "helm": {},
	"grep": {}, "rg": {}, "find": {}, "locate": {},
	"ls": {}, "dir": {}, "cd": {}, "pwd": {},
	"rm": {}, "mv": {}, "cp": {}, "cat": {}, "head": {}, "tail": {},
	"curl": {}, "wget": {}, "make": {}, "go": {}, "npm": {}, "node": {},
	"python": {}, "python3": {}, "pip": {}, "pip3": {},
	"chmod": {}, "chown": {}, "tar": {}, "zip": {}, "unzip": {},
	"ssh": {}, "scp": {}, "rsync": {}, "sudo": {},
	"which": {}, "where": {}, "echo": {}, "export": {}, "set": {},
	"kill": {}, "ps": {}, "top": {}, "df": {}, "du": {},
	"uname": {}, "whoami": {}, "env": {},
	"ping": {}, "netstat": {}, "ss": {}, "traceroute": {}, "tracert": {},
}

func isShellVerb(token string) bool {
	_, ok := shellVerbs[strings.ToLower(token)]
	return ok
}

// IsKnownShellVerb reports whether token is a known shell command head (alias collision check).
func IsKnownShellVerb(token string) bool {
	return isShellVerb(token)
}
