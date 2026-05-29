package openai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/providers"
)

func TestNewClientNoNetwork(t *testing.T) {
	t.Parallel()
	c, err := NewClient("sk-test", "gpt-4.1-mini", "", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if c.http == nil {
		t.Fatal("expected http client")
	}
}

func TestClientChatHappyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "clx/") {
			t.Fatalf("User-Agent = %q", got)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer sk-") {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Temperature != 0 || req.ResponseFormat == nil {
			t.Fatalf("req = %+v", req)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: `{"intent":"find_file","params":{"filename":"*.log"},"confidence":0.9}`}}},
		})
	}))
	defer srv.Close()

	c, err := NewClient("sk-test", "gpt-4.1-mini", srv.URL+"/v1", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.Chat(context.Background(), "sys", "user", map[string]any{"type": "object"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Intent != "find_file" || out.Params["filename"] != "*.log" {
		t.Fatalf("out = %+v", out)
	}
}

func TestClientChatHTTP500(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, _ := NewClient("sk-test", "m", srv.URL+"/v1", time.Second)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatHTTP401(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	c, _ := NewClient("sk-test", "m", srv.URL+"/v1", time.Second)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errInvalidResp) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatEmptyIntent(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: `{"intent":"","params":{}}`}}},
		})
	}))
	defer srv.Close()
	c, _ := NewClient("sk-test", "m", srv.URL+"/v1", time.Second)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errNoMatch) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatBoundedRead(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, io.LimitReader(strings.NewReader(`{"choices":[{"message":{"content":"x"}}]}`), maxResponseBytes+1))
	}))
	defer srv.Close()
	c, _ := NewClient("sk-test", "m", srv.URL+"/v1", time.Second)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errInvalidResp) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatServerDown(t *testing.T) {
	t.Parallel()
	c, err := NewClient("sk-test", "m", "http://127.0.0.1:1/v1", 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatTimeout(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c, _ := NewClient("sk-test", "m", srv.URL+"/v1", 50*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Chat(ctx, "s", "u", nil)
	if !errors.Is(err, errTimeout) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientExplainChatHappyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.ResponseFormat != nil {
			t.Fatal("explain chat must not use response_format")
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "Shows git working tree status."}}},
		})
	}))
	defer srv.Close()
	c, err := NewClient("sk-test", "m", srv.URL+"/v1", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	text, err := c.ExplainChat(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Shows git working tree status." {
		t.Fatalf("text = %q", text)
	}
}

func TestRedactBodyStripsSensitivePatterns(t *testing.T) {
	t.Parallel()
	body := []byte(`{"error":{"message":"Invalid api_key=sk-secret1234567890abcdef"}}`)
	redacted := providers.RedactHTTPBody(body)
	if strings.Contains(redacted, "sk-secret1234567890abcdef") {
		t.Fatalf("secret leaked: %q", redacted)
	}
}
