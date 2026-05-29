package main

import "strings"

// parseConfigSetArgs extracts --stdin and --config from anywhere in args.
// Go's flag package stops at the first positional argument, so flags after the
// path would otherwise be treated as values.
func parseConfigSetArgs(args []string) (configPath string, useStdin bool, positional []string, errMsg string) {
	positional = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--stdin":
			useStdin = true
		case arg == "--config":
			if i+1 >= len(args) {
				return "", false, nil, "config set: --config requires a path"
			}
			i++
			configPath = args[i]
		case strings.HasPrefix(arg, "--config="):
			configPath = strings.TrimPrefix(arg, "--config=")
		default:
			positional = append(positional, arg)
		}
	}
	return configPath, useStdin, positional, ""
}
