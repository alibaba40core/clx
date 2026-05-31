package gemini

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
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	maxResponseBytes = 64 * 1024
)

// Client talks to the Google Gemini generateContent API.
// Construction does not dial the server.
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
		return nil, fmt.Errorf("gemini: api_key is required")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("gemini: model is required")
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

// chatRequest is the Gemini generateContent request body.
type chatRequest struct {
	Contents          []content        `json:"contents"`
	SystemInstruction *content         `json:"systemInstruction,omitempty"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	Temperature      float64        `json:"temperature"`
	ResponseMimeType string         `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]any `json:"responseSchema,omitempty"`
}

// chatResponse is the Gemini generateContent response.
type chatResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content content `json:"content"`
}

// ChatResult is the parsed intent payload from the response.
type ChatResult struct {
	Intent     string
	Params     map[string]string
	Confidence float64
}

// Chat sends a schema-constrained generateContent request and parses the intent JSON.
func (c *Client) Chat(ctx context.Context, system, user string, schema map[string]any) (*ChatResult, error) {
	body, err := json.Marshal(chatRequest{
		Contents: []content{
			{Role: "user", Parts: []part{{Text: user}}},
		},
		SystemInstruction: &content{
			Parts: []part{{Text: system}},
		},
		GenerationConfig: generationConfig{
			Temperature:      0,
			ResponseMimeType: "application/json",
			ResponseSchema:   schema,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: encode request: %w", err)
	}

	data, err := c.doPost(ctx, body)
	if err != nil {
		return nil, err
	}

	var resp chatResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errInvalidResp
	}
	text := extractText(resp)
	if text == "" {
		return nil, errNoMatch
	}

	return parseIntentContent(text)
}

// ExplainChat sends a plain-text generateContent request for command explanation.
func (c *Client) ExplainChat(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Contents: []content{
			{Role: "user", Parts: []part{{Text: user}}},
		},
		SystemInstruction: &content{
			Parts: []part{{Text: system}},
		},
		GenerationConfig: generationConfig{
			Temperature: 0,
		},
	})
	if err != nil {
		return "", fmt.Errorf("gemini: encode request: %w", err)
	}

	data, err := c.doPost(ctx, body)
	if err != nil {
		return "", err
	}

	var resp chatResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errInvalidResp
	}
	text := strings.TrimSpace(extractText(resp))
	if text == "" {
		return "", errNoMatch
	}
	return text, nil
}

// CommandChat sends a schema-constrained generateContent request for command generation.
func (c *Client) CommandChat(ctx context.Context, system, user string, schema map[string]any) (*providers.CommandResponse, error) {
	body, err := json.Marshal(chatRequest{
		Contents: []content{
			{Role: "user", Parts: []part{{Text: user}}},
		},
		SystemInstruction: &content{
			Parts: []part{{Text: system}},
		},
		GenerationConfig: generationConfig{
			Temperature:      0,
			ResponseMimeType: "application/json",
			ResponseSchema:   schema,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: encode request: %w", err)
	}

	data, err := c.doPost(ctx, body)
	if err != nil {
		return nil, err
	}

	var resp chatResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, errInvalidResp
	}
	text := extractText(resp)
	if text == "" {
		return nil, errNoMatch
	}

	out, err := providers.ParseCommandContent(text)
	if err != nil {
		return nil, mapParseError(err)
	}
	return out, nil
}

// doPost sends the JSON body to the generateContent endpoint and returns the raw response.
func (c *Client) doPost(ctx context.Context, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, c.model)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "clx/"+cliversion.Version)
	// Pass the API key via header rather than the URL query string to keep it
	// out of request URLs, proxy logs, and any *url.Error strings.
	req.Header.Set("x-goog-api-key", c.apiKey)

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

	if err := providers.HTTPStatusError(resp.StatusCode, "gemini", data, c.logger); err != nil {
		return nil, err
	}
	return data, nil
}

func extractText(resp chatResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}
	parts := resp.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0].Text)
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

// mapParseError converts the public sentinels returned by
// providers.ParseCommandContent back into the client's internal sentinels so the
// provider's mapClientError performs the single, consistent public mapping.
func mapParseError(err error) error {
	switch {
	case errors.Is(err, providers.ErrInvalidResp):
		return errInvalidResp
	case errors.Is(err, providers.ErrNoMatch):
		return errNoMatch
	default:
		return err
	}
}

func mapRoundTripError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return providers.ErrTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return providers.ErrTimeout
	}
	return providers.ErrUnavailable
}
