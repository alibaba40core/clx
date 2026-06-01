package executor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
	"github.com/alibaba40core/clx/internal/shellchain"
)

const (
	maxChainStages   = 8
	maxChainExpr     = 4
	maxExprTokenBytes = 256
	maxStageTokens   = 64
	exprForbiddenChars = ";|&<>`\n\r"
	exprForbiddenCmd   = "%"
)

var (
	ErrChainEmpty    = errors.New("executor: command chain has no stages")
	ErrChainTooLong  = errors.New("executor: command chain has too many stages")
	ErrChainExprCap  = errors.New("executor: command chain has too many expression tokens")
)

// ValidateExpressionToken checks a constrained scriptblock/predicate token.
func ValidateExpressionToken(tok string, shell string) error {
	if tok == "" {
		return fmt.Errorf("%w: empty expression token", ErrAIArgvToken)
	}
	if len(tok) > maxExprTokenBytes {
		return fmt.Errorf("%w: expression token too long (%d bytes)", ErrAIArgvToken, len(tok))
	}
	if strings.ContainsRune(tok, 0) {
		return fmt.Errorf("%w: null byte in expression", ErrAIArgvToken)
	}
	bad := exprForbiddenChars
	if strings.EqualFold(shell, "cmd") {
		bad += exprForbiddenCmd
	}
	if strings.ContainsAny(tok, bad) {
		return fmt.Errorf("%w: forbidden characters in expression %q", ErrAIArgvToken, tok)
	}
	return nil
}

// ValidateCommandChain enforces safety on an UNTRUSTED command chain.
func ValidateCommandChain(chain *generator.CommandChain, shell string) error {
	if chain == nil || len(chain.Stages) == 0 {
		return ErrChainEmpty
	}
	if len(chain.Stages) < 2 {
		return fmt.Errorf("%w: chain needs at least 2 stages", ErrChainEmpty)
	}
	if len(chain.Stages) > maxChainStages {
		return fmt.Errorf("%w: %d (cap %d)", ErrChainTooLong, len(chain.Stages), maxChainStages)
	}
	chain.NormalizeConnectors()
	if len(chain.Connectors) != len(chain.Stages)-1 {
		return fmt.Errorf("%w: connector count mismatch", ErrChainEmpty)
	}
	exprCount := 0
	for _, st := range chain.Stages {
		if len(st.Tokens) == 0 {
			return fmt.Errorf("%w: empty stage", ErrAIArgvToken)
		}
		if len(st.Tokens) > maxStageTokens {
			return fmt.Errorf("%w: stage has too many tokens", ErrAIArgvToken)
		}
		for _, tok := range st.Tokens {
			if tok.Value == "" {
				return fmt.Errorf("%w: empty token", ErrAIArgvToken)
			}
			if tok.Expr {
				exprCount++
				if err := ValidateExpressionToken(tok.Value, shell); err != nil {
					return err
				}
				continue
			}
			if err := validatePlainToken(tok.Value, shell); err != nil {
				return err
			}
		}
	}
	if exprCount > maxChainExpr {
		return fmt.Errorf("%w: %d (cap %d)", ErrChainExprCap, exprCount, maxChainExpr)
	}
	return nil
}

// BuildValidatedChainScript joins validated stages with shell-native connector symbols.
func BuildValidatedChainScript(shell string, chain *generator.CommandChain, profile environment.SystemProfile) (string, error) {
	if err := ValidateCommandChain(chain, shell); err != nil {
		return "", err
	}
	quote := QuotePOSIX
	switch strings.ToLower(shell) {
	case "powershell", "pwsh":
		quote = QuotePowerShell
	case "cmd":
		quote = QuoteCmd
	}
	stageParts := make([]string, len(chain.Stages))
	for i, st := range chain.Stages {
		var b strings.Builder
		for j, tok := range st.Tokens {
			if j > 0 {
				b.WriteByte(' ')
			}
			if tok.Expr {
				b.WriteString(tok.Value)
			} else {
				b.WriteString(quote(tok.Value))
			}
		}
		stageParts[i] = b.String()
	}
	if len(stageParts) < 2 {
		return "", ErrChainEmpty
	}
	var script strings.Builder
	script.WriteString(stageParts[0])
	for i := 1; i < len(stageParts); i++ {
		sym, err := shellchain.Symbol(chain.Connectors[i-1], profile)
		if err != nil {
			return "", err
		}
		script.WriteByte(' ')
		script.WriteString(sym)
		script.WriteByte(' ')
		script.WriteString(stageParts[i])
	}
	out := script.String()
	if out == "" {
		return "", ErrEmptyScriptArgv
	}
	return out, nil
}
