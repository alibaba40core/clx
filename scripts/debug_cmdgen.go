//go:build ignore

// One-off diagnostic: go run scripts/debug_cmdgen.go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alibaba40core/clx/internal/config"
	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/providers"
	"github.com/alibaba40core/clx/internal/providers/factory"
)

func main() {
	raw := "remove xyz file from my dir"
	if len(os.Args) > 1 {
		raw = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	_, err := config.Bootstrap(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap: %v\n", err)
		os.Exit(1)
	}
	cfgPath, err := config.ConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config path: %v\n", err)
		os.Exit(1)
	}
	cfg, err := config.Load(ctx, cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("provider=%q openai.model=%q\n", cfg.Provider, cfg.Providers.OpenAI.Model)

	prof, err := environment.LoadOrDetect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "profile: %v\n", err)
		os.Exit(1)
	}

	p, err := factory.NewFromConfig(cfg, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "factory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("provider name=%q\n", p.Name())

	cg, ok := p.(providers.CommandGenerator)
	if !ok {
		fmt.Fprintln(os.Stderr, "provider does not implement CommandGenerator")
		os.Exit(1)
	}

	resp, err := cg.GenerateCommand(ctx, providers.CommandRequest{RawInput: raw, Profile: prof})
	if err != nil {
		fmt.Printf("error: %T %v\n", err, err)
		fmt.Printf("  ErrNoMatch=%v ErrInvalidResp=%v ErrAuth=%v ErrUnavailable=%v ErrTimeout=%v\n",
			errors.Is(err, providers.ErrNoMatch),
			errors.Is(err, providers.ErrInvalidResp),
			errors.Is(err, providers.ErrAuth),
			errors.Is(err, providers.ErrUnavailable),
			errors.Is(err, providers.ErrTimeout),
		)
		fmt.Fprintln(os.Stderr, "hint: CLX_LOGGING_LEVEL=debug clx ... logs openai content_preview on parse failure")
		os.Exit(1)
	}
	fmt.Printf("ok: argv=%v chain=%v shell=%q conf=%.2f\n", resp.Argv, resp.HasChain(), resp.Shell, resp.Confidence)
}
