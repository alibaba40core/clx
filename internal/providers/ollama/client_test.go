package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/providers"
)

func TestNewClientNoNetwork(t *testing.T) {
	t.Parallel()
	c, err := NewClient("http://127.0.0.1:1", "qwen3:4b", 2*time.Second, nil)
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
		if r.URL.Path != "/api/chat" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "clx/") {
			t.Fatalf("User-Agent = %q", got)
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Stream || req.Options.Temperature != 0 {
			t.Fatalf("req = %+v", req)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Message: struct {
				Content string `json:"content"`
			}{Content: `{"intent":"find_file","params":{"filename":"*.log"},"confidence":0.9}`},
		})
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "qwen3:4b", 5*time.Second, nil)
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
	c, _ := NewClient(srv.URL, "m", time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatHTTP400(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, "m", time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrInvalidResp) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatEmptyIntent(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(chatResponse{
			Message: struct {
				Content string `json:"content"`
			}{Content: `{"intent":"","params":{}}`},
		})
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, "m", time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errNoMatch) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatBoundedRead(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, io.LimitReader(strings.NewReader(`{"message":{"content":"x"}}`), maxResponseBytes+1))
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, "m", time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	// Truncated JSON should fail parse → ErrInvalidResp
	if !errors.Is(err, errInvalidResp) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientChatServerDown(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	c, err := NewClient("http://"+addr, "m", time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrUnavailable) {
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
	c, _ := NewClient(srv.URL, "m", 50*time.Millisecond, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Chat(ctx, "s", "u", nil)
	if !errors.Is(err, providers.ErrTimeout) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientExplainChatHappyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Format != nil {
			t.Fatal("explain chat must not use format")
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Message: struct {
				Content string `json:"content"`
			}{Content: "Lists files in the current directory."},
		})
	}))
	defer srv.Close()
	c, err := NewClient(srv.URL, "m", 5*time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	text, err := c.ExplainChat(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Lists files in the current directory." {
		t.Fatalf("text = %q", text)
	}
}

func TestClientExplainChatEmpty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(chatResponse{
			Message: struct {
				Content string `json:"content"`
			}{Content: "   "},
		})
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, "m", time.Second, nil)
	_, err := c.ExplainChat(context.Background(), "s", "u")
	if !errors.Is(err, errNoMatch) {
		t.Fatalf("err = %v", err)
	}
}
