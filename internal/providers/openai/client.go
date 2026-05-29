package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/cliversion"
	"github.com/alibaba40core/clx/internal/providers"
)

const (
	defaultBaseURL     = "https://api.openai.com/v1"
	maxResponseBytes   = 64 * 1024
	intentResponseName = "clx_intent"
)

// Client talks to the OpenAI chat completions API. Construction does not dial the server.
type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
	logger  *slog.Logger
}

// NewClient returns a client for apiKey and model. baseURL may be empty for the default.
func NewClient(apiKey, model, baseURL string, timeout time.Duration, logger *slog.Logger) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("openai: api_key is required")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("openai: model is required")
	}
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = defaultBaseURL
	}
	base = strings.TrimRight(base, "/")
	return &Client{
		baseURL: base,
		apiKey:  apiKey,
		model:   model,
		http: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}, nil
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ChatResult is the parsed intent payload from message content.
type ChatResult struct {
	Intent     string
	Params     map[string]string
	Confidence float64
}

// Chat sends a schema-constrained chat completion and parses the intent JSON.
func (c *Client) Chat(ctx context.Context, system, user string, schema map[string]any) (*ChatResult, error) {
	responseFormat := map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   intentResponseName,
			"strict": true,
			"schema": schema,
		},
	}
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:    0,
		ResponseFormat: responseFormat,
	})
	if err != nil {
		return nil, fmt.Errorf("openai: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "clx/"+cliversion.Version)

	resp, err := c.http.Do(req)
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
		providers.DebugLogHTTPError(c.logger, "openai", resp.StatusCode, data)
		return nil, errInvalidResp
	}

	var chat chatResponse
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, errInvalidResp
	}
	if len(chat.Choices) == 0 || strings.TrimSpace(chat.Choices[0].Message.Content) == "" {
		return nil, errNoMatch
	}

	return parseIntentContent(chat.Choices[0].Message.Content)
}

// ExplainChat sends a plain-text chat completion for command explanation (Phase 2.4).
func (c *Client) ExplainChat(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("openai: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "clx/"+cliversion.Version)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", mapRoundTripError(err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", errInvalidResp
	}

	if resp.StatusCode >= 500 {
		return "", errUnavailable
	}
	if resp.StatusCode >= 400 {
		providers.DebugLogHTTPError(c.logger, "openai", resp.StatusCode, data)
		return "", errInvalidResp
	}

	var chat chatResponse
	if err := json.Unmarshal(data, &chat); err != nil {
		return "", errInvalidResp
	}
	if len(chat.Choices) == 0 {
		return "", errNoMatch
	}
	text := strings.TrimSpace(chat.Choices[0].Message.Content)
	if text == "" {
		return "", errNoMatch
	}
	return text, nil
}

type parsedIntent struct {
	Intent     string            `json:"intent"`
	Params     map[string]string `json:"params"`
	Confidence float64           `json:"confidence"`
}

func parseIntentContent(content string) (*ChatResult, error) {
	var parsed parsedIntent
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, errInvalidResp
	}
	if strings.TrimSpace(parsed.Intent) == "" {
		return nil, errNoMatch
	}
	if parsed.Params == nil {
		parsed.Params = map[string]string{}
	}
	return &ChatResult{
		Intent:     parsed.Intent,
		Params:     parsed.Params,
		Confidence: parsed.Confidence,
	}, nil
}

func mapRoundTripError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return errTimeout
	}
	return errUnavailable
}

