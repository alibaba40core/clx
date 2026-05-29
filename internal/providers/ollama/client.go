package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba40core/clx/internal/cliversion"
)

const maxResponseBytes = 64 * 1024

// Client talks to a local Ollama daemon over HTTP. Construction does not dial the server.
type Client struct {
	baseURL string
	model   string
	http    *http.Client
}

// NewClient returns a client for host (e.g. http://localhost:11434) and model name.
func NewClient(host, model string, timeout time.Duration) (*Client, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("ollama: host is required")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("ollama: model is required")
	}
	base := strings.TrimRight(host, "/")
	return &Client{
		baseURL: base,
		model:   model,
		http: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// ChatRequest is the payload for POST /api/chat.
type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []chatMessage  `json:"messages"`
	Stream   bool           `json:"stream"`
	Format   map[string]any `json:"format,omitempty"`
	Options  chatOptions    `json:"options"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatOptions struct {
	Temperature float64 `json:"temperature"`
}

// chatResponse is a minimal subset of Ollama's /api/chat response.
type chatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// ChatResult is the parsed intent payload from Ollama message.content.
type ChatResult struct {
	Intent     string
	Params     map[string]string
	Confidence float64
}

// Chat sends a schema-constrained chat completion and parses the intent JSON.
func (c *Client) Chat(ctx context.Context, system, user string, format map[string]any) (*ChatResult, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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
		return nil, errInvalidResp
	}

	var chat chatResponse
	if err := json.Unmarshal(data, &chat); err != nil {
		return nil, errInvalidResp
	}
	if strings.TrimSpace(chat.Message.Content) == "" {
		return nil, errNoMatch
	}

	return parseIntentContent(chat.Message.Content)
}

// ExplainChat sends a plain-text chat completion for command explanation (Phase 2.4).
func (c *Client) ExplainChat(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(ChatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream:  false,
		Options: chatOptions{Temperature: 0},
	})
	if err != nil {
		return "", fmt.Errorf("ollama: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
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
		return "", errInvalidResp
	}

	var chat chatResponse
	if err := json.Unmarshal(data, &chat); err != nil {
		return "", errInvalidResp
	}
	text := strings.TrimSpace(chat.Message.Content)
	if text == "" {
		return "", errNoMatch
	}
	return text, nil
}

// parsedIntent is the JSON shape we expect inside message.content.
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
