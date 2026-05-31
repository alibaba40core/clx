package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alibaba40core/clx/internal/cliversion"
	"github.com/alibaba40core/clx/internal/providers"
)

// GenerateCommand builds the command prompt/schema and calls OpenAI, returning a
// provider-generated argv command. The result is untrusted and must be validated
// and gated by the caller before execution.
func (p *Provider) GenerateCommand(ctx context.Context, req providers.CommandRequest) (*providers.CommandResponse, error) {
	system, user, err := providers.BuildCommandPrompt(req)
	if err != nil {
		return nil, err
	}
	schema := providers.BuildCommandSchema()
	out, err := p.client.CommandChat(ctx, system, user, schema)
	if err != nil {
		return nil, mapClientError(err)
	}
	return out, nil
}

// CommandChat sends a schema-constrained chat completion and parses the command JSON.
func (c *Client) CommandChat(ctx context.Context, system, user string, schema map[string]any) (*providers.CommandResponse, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:    0,
		ResponseFormat: providers.BuildOpenAICommandResponseFormat(schema),
	})
	if err != nil {
		return nil, fmt.Errorf("openai: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("User-Agent", "clx/"+cliversion.Version)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, mapRoundTripError(err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, errInvalidResp
	}

	if err := providers.HTTPStatusError(resp.StatusCode, "openai", data, c.logger); err != nil {
		return nil, err
	}

	var chat chatResponse
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, errInvalidResp
	}
	if len(chat.Choices) == 0 || strings.TrimSpace(chat.Choices[0].Message.Content) == "" {
		return nil, errNoMatch
	}
	return providers.ParseCommandContent(chat.Choices[0].Message.Content)
}
