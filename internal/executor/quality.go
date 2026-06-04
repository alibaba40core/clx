package executor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/generator"
)

var ErrCommandQuality = errors.New("executor: ai command quality rejected")

var placeholderTokens = map[string]struct{}{
	"url":             {},
	"file":            {},
	"linkname":        {},
	"targetdirectory": {},
	"file.tar.gz":     {},
}

// ValidateCommandQuality rejects common low-quality AI outputs after safety validation.
func ValidateCommandQuality(gen generator.GeneratedCommand, rawInput string) error {
	shell := strings.ToLower(strings.TrimSpace(gen.Shell))
	if gen.Chain != nil {
		if err := validateChainQuality(gen.Chain, rawInput, shell); err != nil {
			return err
		}
	}
	if len(gen.Argv) > 0 {
		if err := validateArgvQuality(gen.Argv, rawInput); err != nil {
			return err
		}
	}
	if err := validateCompareIntent(rawInput, gen); err != nil {
		return err
	}
	return nil
}

func validateArgvQuality(argv []string, rawInput string) error {
	for i, tok := range argv {
		prev := ""
		if i > 0 {
			prev = argv[i-1]
		}
		if isPlaceholderToken(tok, prev, rawInput) {
			return fmt.Errorf("%w: placeholder token %q", ErrCommandQuality, tok)
		}
	}
	return nil
}

func validateChainQuality(chain *generator.CommandChain, rawInput string, shell string) error {
	if chain == nil {
		return nil
	}
	for _, st := range chain.Stages {
		for i, tok := range st.Tokens {
			prev := ""
			if i > 0 {
				prev = st.Tokens[i-1].Value
			}
			if isPlaceholderToken(tok.Value, prev, rawInput) {
				return fmt.Errorf("%w: placeholder token %q", ErrCommandQuality, tok.Value)
			}
		}
	}
	if shell != "powershell" {
		return nil
	}
	chain.NormalizeConnectors()
	for i := 1; i < len(chain.Stages); i++ {
		if i-1 >= len(chain.Connectors) || chain.Connectors[i-1] != generator.ChainPipe {
			continue
		}
		st := chain.Stages[i]
		if len(st.Tokens) == 0 {
			continue
		}
		cmd := strings.TrimSpace(st.Tokens[0].Value)
		if cmd == "" {
			continue
		}
		if isBrokenFilterStage(cmd, st) {
			return fmt.Errorf("%w: pipe stage must use Where-Object or ForEach-Object, not bare %q", ErrCommandQuality, cmd)
		}
	}
	_ = rawInput
	return nil
}

func isBrokenFilterStage(cmd string, st generator.ChainStage) bool {
	if cmd == "$_" || strings.HasPrefix(cmd, "$_") {
		return true
	}
	if strings.HasPrefix(cmd, "{") && !isFilterCmdlet(cmd) {
		return true
	}
	for _, tok := range st.Tokens {
		if tok.Expr {
			continue
		}
		v := strings.TrimSpace(tok.Value)
		if v == "$_" {
			return true
		}
	}
	return false
}

func isFilterCmdlet(cmd string) bool {
	switch strings.ToLower(cmd) {
	case "where-object", "foreach-object", "select-object":
		return true
	default:
		return false
	}
}

var pathFlagTokens = map[string]struct{}{
	"-path":              {},
	"-literalpath":       {},
	"-destinationpath":   {},
	"-c":                 {},
}

var dotAllowedInputHints = []string{
	"folder", "directory", "this", "current", "tree", "project", "archive",
	"compress", "zip", "largest", "todo", "source", "empty", "duplicate",
	"recycle", "downloads", "here", "workspace",
}

func isPlaceholderToken(tok, prevTok, rawInput string) bool {
	v := strings.TrimSpace(tok)
	if v == "" {
		return false
	}
	if _, ok := placeholderTokens[strings.ToLower(v)]; ok {
		return true
	}
	if v != "." {
		return false
	}
	if _, ok := pathFlagTokens[strings.ToLower(strings.TrimSpace(prevTok))]; ok {
		return false
	}
	lower := strings.ToLower(rawInput)
	for _, hint := range dotAllowedInputHints {
		if strings.Contains(lower, hint) {
			return false
		}
	}
	return true
}

func validateCompareIntent(rawInput string, gen generator.GeneratedCommand) error {
	lower := strings.ToLower(rawInput)
	if !strings.Contains(lower, "compare") {
		return nil
	}
	if !strings.Contains(lower, "folder") && !strings.Contains(lower, "directory") {
		return nil
	}
	if hasCompareObject(gen) {
		return nil
	}
	paths := distinctPathLikeTokens(gen)
	if len(paths) >= 2 {
		return nil
	}
	return fmt.Errorf("%w: compare folders requires Compare-Object or two distinct paths", ErrCommandQuality)
}

func hasCompareObject(gen generator.GeneratedCommand) bool {
	if gen.Chain != nil {
		for _, st := range gen.Chain.Stages {
			for _, tok := range st.Tokens {
				if strings.EqualFold(strings.TrimSpace(tok.Value), "Compare-Object") {
					return true
				}
			}
		}
	}
	for _, tok := range gen.Argv {
		if strings.EqualFold(strings.TrimSpace(tok), "Compare-Object") {
			return true
		}
	}
	return false
}

func distinctPathLikeTokens(gen generator.GeneratedCommand) []string {
	seen := make(map[string]struct{}, 4)
	var out []string
	add := func(v string) {
		v = strings.TrimSpace(v)
		if !pathLikeToken(v) {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if gen.Chain != nil {
		for _, st := range gen.Chain.Stages {
			for _, tok := range st.Tokens {
				if tok.Expr {
					continue
				}
				add(tok.Value)
			}
		}
	}
	for _, tok := range gen.Argv {
		add(tok)
	}
	return out
}

func pathLikeToken(v string) bool {
	if v == "" || isPlaceholderToken(v, "", "") {
		return false
	}
	if strings.ContainsAny(v, `/\`) {
		return true
	}
	return false
}
