package ollama

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

// GenerateCommand builds the command prompt/schema and calls Ollama /api/chat,
// returning a provider-generated argv command. The result is untrusted and must
// be validated and gated by the caller before execution.
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
func (c *Client) CommandChat(ctx context.Context, system, user string, format map[string]any) (*providers.CommandResponse, error) {
	body, err := json.Marshal(ChatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream:  false,
		Format:  format,
		Options: chatOptions{Temperature: 0},
	})
	if err != nil {
		return nil, fmt.Errorf("ollama: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
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

	if resp.StatusCode >= 500 {
		return nil, errUnavailable
	}
	if resp.StatusCode >= 400 {
		providers.DebugLogHTTPError(c.logger, "ollama", resp.StatusCode, data)
		return nil, errInvalidResp
	}

	var chat chatResponse
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, errInvalidResp
	}
	if strings.TrimSpace(chat.Message.Content) == "" {
		return nil, errNoMatch
	}
	return providers.ParseCommandContent(chat.Message.Content)
}
