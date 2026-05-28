package intent

import (
	"context"

	"github.com/alibaba40core/clx/internal/parser"
)

// Resolver maps a parsed request to a resolved intent.
// Implementations MUST return ErrNotFound (not nil intent, not other error)
// to signal "I don't know, try the next resolver."
type Resolver interface {
	Resolve(ctx context.Context, req parser.Request) (ResolvedIntent, error)
}
