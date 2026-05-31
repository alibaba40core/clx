package main

import (
	"context"
	"fmt"
	"io"

	"github.com/alibaba40core/clx/internal/cache"
)

func runCache(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printCacheHelp(stderr)
		return 2
	}

	switch args[0] {
	case "status":
		return runCacheStatus(stdout, stderr)
	case "clear":
		return runCacheClear(stdout, stderr)
	case "help", "--help", "-h":
		printCacheHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "cache: unknown subcommand %q\n", args[0])
		printCacheHelp(stderr)
		return 2
	}
}

func printCacheHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  clx cache status")
	fmt.Fprintln(w, "  clx cache clear")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Inspect or reset local caches (~/.clx/cache/): intents, explanations, commands.")
	fmt.Fprintln(w, "Intent cache: AI-resolved intents. Commands cache: AI command-generation argv.")
	fmt.Fprintln(w, "Enable caching with: clx config set features.cache_commands true")
}

func runCacheStatus(stdout, stderr io.Writer) int {
	ctx := context.Background()
	cfg, code := loadConfigForEdit("", stderr)
	if code != 0 {
		return code
	}
	if !cfg.Features.CacheCommands {
		fmt.Fprintln(stderr, "note: features.cache_commands is false; pipeline will not read/write caches")
	}
	stats, err := cache.AllStats(ctx, cfg.Cache, nil)
	if err != nil {
		fmt.Fprintf(stderr, "cache status: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "CLX cache:")
	for _, s := range stats {
		fmt.Fprintf(stdout, "  %s: %d entries, %d bytes\n", s.Name, s.Entries, s.Bytes)
		fmt.Fprintf(stdout, "    path: %s\n", s.Path)
	}
	return 0
}

func runCacheClear(stdout, stderr io.Writer) int {
	ctx := context.Background()
	cfg, code := loadConfigForEdit("", stderr)
	if code != 0 {
		return code
	}
	if !cfg.Features.CacheCommands {
		fmt.Fprintln(stderr, "note: features.cache_commands is false; clear still removes on-disk cache files")
	}
	if err := cache.ClearAll(ctx, cfg.Cache, nil); err != nil {
		fmt.Fprintf(stderr, "cache clear: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "cache cleared (intents + explanations + commands)")
	return 0
}
