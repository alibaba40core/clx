package executor

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrAIArgvEmpty is returned when AI produces no usable argv.
	ErrAIArgvEmpty = errors.New("executor: ai command has no argv")
	// ErrAIArgvTooLong is returned when AI argv exceeds the token cap.
	ErrAIArgvTooLong = errors.New("executor: ai command has too many tokens")
	// ErrAIArgvToken is returned for an individual unsafe argv token.
	ErrAIArgvToken = errors.New("executor: ai command token rejected")
)

const (
	// maxAIArgvTokens bounds the number of argv elements an AI command may have.
	maxAIArgvTokens = 64
	// maxAIArgvTokenBytes bounds the length of a single argv element.
	maxAIArgvTokenBytes = 4096
)

// ValidateGeneratedArgv enforces the safety invariants on an UNTRUSTED,
// AI-generated argv before it is risk-assessed and executed. It guarantees the
// command is a single program invocation with no shell operators, so it can be
// run argv-only. This is the trust boundary for AI-generated commands and is
// applied in addition to the exec-time BuildValidatedScript check (defense in
// depth). It does NOT decide whether the command is allowed — that is risk +
// policy — only that it is well-formed and free of injection metacharacters.
func ValidateGeneratedArgv(argv []string, shell string) error {
	if len(argv) == 0 {
		return ErrAIArgvEmpty
	}
	if len(argv) > maxAIArgvTokens {
		return fmt.Errorf("%w: %d (cap %d)", ErrAIArgvTooLong, len(argv), maxAIArgvTokens)
	}
	bad := metacharsForShell(shell)
	for _, tok := range argv {
		if tok == "" {
			return fmt.Errorf("%w: empty token", ErrAIArgvToken)
		}
		if len(tok) > maxAIArgvTokenBytes {
			return fmt.Errorf("%w: token too long (%d bytes)", ErrAIArgvToken, len(tok))
		}
		if strings.ContainsRune(tok, 0) {
			return fmt.Errorf("%w: null byte in %q", ErrAIArgvToken, tok)
		}
		if strings.ContainsAny(tok, bad) {
			return fmt.Errorf("%w: shell metacharacters in %q", ErrAIArgvToken, tok)
		}
	}
	return nil
}
