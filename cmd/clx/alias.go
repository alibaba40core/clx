package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/alibaba40core/clx/internal/aliases"
	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/intent"
)

func runAlias(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printAliasHelp(stderr)
		return 2
	}
	switch args[0] {
	case "set":
		return runAliasSet(args[1:], stdout, stderr)
	case "list":
		return runAliasList(args[1:], stdout, stderr)
	case "rm", "remove":
		return runAliasRm(args[1:], stdout, stderr)
	case "help", "--help", "-h":
		printAliasHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "alias: unknown subcommand %q\n", args[0])
		printAliasHelp(stderr)
		return 2
	}
}

func printAliasHelp(w io.Writer) {
	fmt.Fprintln(w, "CLX alias — user-global command shortcuts")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx alias set <name> \"<value>\" [--force] [--config path]")
	fmt.Fprintln(w, "  clx alias list [--config path]")
	fmt.Fprintln(w, "  clx alias rm <name> [--config path]")
	fmt.Fprintln(w, "  clx alias help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Aliases are stored in ~/.clx/aliases.yaml. The first token of input is expanded")
	fmt.Fprintln(w, "at parse time (single level only). Expanded commands still pass risk, policy,")
	fmt.Fprintln(w, "and safety gates like any other clx invocation.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  clx alias set gst \"git status\"")
	fmt.Fprintln(w, "  clx alias list")
	fmt.Fprintln(w, "  clx --explain gst")
	fmt.Fprintln(w, "  clx -y gst")
}

func runAliasSet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("alias set", flag.ContinueOnError)
	fs.SetOutput(stderr)
	force := fs.Bool("force", false, "set alias even when name collides with shell verbs or rule examples")
	forceShort := fs.Bool("f", false, "alias for --force")
	configPath := fs.String("config", "", "path to config.yaml")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 2 {
		fmt.Fprintln(stderr, "alias set: name and value required")
		return 2
	}
	name := fs.Arg(0)
	value := strings.Join(fs.Args()[1:], " ")

	path := *configPath
	if path == "" {
		var code int
		path, code = defaultConfigPath(stderr)
		if code != 0 {
			return code
		}
	}
	cfg, code := loadConfigForEdit(path, stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	if _, err := config.Bootstrap(ctx); err != nil {
		fmt.Fprintf(stderr, "bootstrap: %v\n", err)
		return 1
	}
	store, err := aliases.Open(ctx, cfg.Aliases.MaxAliases)
	if err != nil {
		fmt.Fprintf(stderr, "aliases: %v\n", err)
		return 1
	}
	eng, err := intent.NewEngineWithOverlay(ctx, nil)
	if err != nil {
		fmt.Fprintf(stderr, "rules: %v\n", err)
		return 1
	}
	doForce := *force || *forceShort
	for _, w := range aliases.CollisionWarnings(name, eng) {
		fmt.Fprintf(stderr, "warning: %s\n", w)
		if !doForce {
			fmt.Fprintln(stderr, "use --force to set anyway")
			return 2
		}
	}
	if err := store.Set(ctx, name, value); err != nil {
		fmt.Fprintf(stderr, "alias set: %v\n", err)
		return 1
	}
	aliasPath, _ := config.AliasesPath()
	if aliasPath != "" {
		fmt.Fprintf(stdout, "alias %q set (saved to %s)\n", name, aliasPath)
	} else {
		fmt.Fprintf(stdout, "alias %q set\n", name)
	}
	return 0
}

func runAliasList(args []string, stdout, stderr io.Writer) int {
	configPath, code := parseConfigPathFlag(args, stderr)
	if code != 0 {
		return code
	}
	cfg, code := loadConfigForEdit(configPath, stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	store, err := aliases.Open(ctx, cfg.Aliases.MaxAliases)
	if err != nil {
		fmt.Fprintf(stderr, "aliases: %v\n", err)
		return 1
	}
	list := store.List()
	if len(list) == 0 {
		fmt.Fprintln(stdout, "no aliases defined")
		return 0
	}
	for _, e := range list {
		fmt.Fprintf(stdout, "%s\t%s\n", e.Name, e.Value)
	}
	return 0
}

func runAliasRm(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("alias rm", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "path to config.yaml")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "alias rm: name required")
		return 2
	}
	name := fs.Arg(0)
	path := *configPath
	if path == "" {
		var code int
		path, code = defaultConfigPath(stderr)
		if code != 0 {
			return code
		}
	}
	cfg, code := loadConfigForEdit(path, stderr)
	if code != 0 {
		return code
	}
	ctx := context.Background()
	store, err := aliases.Open(ctx, cfg.Aliases.MaxAliases)
	if err != nil {
		fmt.Fprintf(stderr, "aliases: %v\n", err)
		return 1
	}
	if err := store.Remove(ctx, name); err != nil {
		fmt.Fprintf(stderr, "alias rm: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "alias %q removed\n", name)
	return 0
}

func defaultConfigPath(stderr io.Writer) (string, int) {
	path, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(stderr, "config path: %v\n", err)
		return "", 1
	}
	return path, 0
}
