package providers

import (
	"context"
	"errors"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/generator"
)

// Provider resolves natural-language input to a structured intent via an LLM.
// Implementations must stay stateless and must not import internal/memory.
type Provider interface {
	Name() string
	ResolveIntent(ctx context.Context, req IntentRequest) (*IntentResponse, error)
	Explain(ctx context.Context, gen generator.GeneratedCommand) (string, error)
}

// IntentRequest carries everything the prompt builder and schema builder need.
type IntentRequest struct {
	RawInput     string
	Profile      environment.SystemProfile
	KnownIntents []string            // closed vocabulary, sorted, capped at 256
	RuleParams   map[string][]string // intent name -> declared param names
	SkillHints   map[string]string   // skill pack name -> optional domain prompt
}

// IntentResponse is a provider-level resolved intent before Source is stamped.
type IntentResponse struct {
	Intent     string
	Params     map[string]string
	Confidence float64
}

// Sentinel errors for provider implementations and the intent adapter.
var (
	ErrUnavailable = errors.New("provider unavailable")
	ErrTimeout     = errors.New("provider timeout")
	ErrInvalidResp = errors.New("provider invalid response")
	ErrNoMatch     = errors.New("provider no match")
)
