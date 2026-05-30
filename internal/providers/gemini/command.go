package gemini

import (
	"context"

	"github.com/alibaba40core/clx/internal/providers"
)

// GenerateCommand builds the command prompt/schema and calls Gemini, returning a
// provider-generated argv command. The result is untrusted and must be validated
// and gated by the caller before execution.
func (p *Provider) GenerateCommand(ctx context.Context, req providers.CommandRequest) (*providers.CommandResponse, error) {
	system, user, err := providers.BuildCommandPrompt(req)
	if err != nil {
		return nil, err
	}
	schema := providers.BuildCommandSchema()
	schema = stripAdditionalProperties(schema)

	out, err := p.client.CommandChat(ctx, system, user, schema)
	if err != nil {
		return nil, mapClientError(err)
	}
	return out, nil
}

var _ providers.CommandGenerator = (*Provider)(nil)
