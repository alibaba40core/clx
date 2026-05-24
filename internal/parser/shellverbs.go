package parser

import "strings"

// shellVerbs is a bounded set of known shell command first tokens.
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
}

func isShellVerb(token string) bool {
	_, ok := shellVerbs[strings.ToLower(token)]
	return ok
}
