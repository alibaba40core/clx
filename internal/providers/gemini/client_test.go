package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/providers"
)

func TestNewClientNoNetwork(t *testing.T) {
	t.Parallel()
	c, err := NewClient("AIza-test-key", "gemini-2.0-flash", "", 2*time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.http == nil {
		t.Fatal("expected http client")
	}
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	t.Parallel()
	_, err := NewClient("", "gemini-2.0-flash", "", time.Second, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientRequiresModel(t *testing.T) {
	t.Parallel()
	_, err := NewClient("AIza-test", "", "", time.Second, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClientChatHappyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "generateContent") {
			t.Fatalf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "clx/") {
			t.Fatalf("User-Agent = %q", got)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "AIza-test" {
			t.Fatalf("x-goog-api-key = %q", got)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.GenerationConfig.Temperature != 0 {
			t.Fatalf("temperature = %v", req.GenerationConfig.Temperature)
		}
		if req.GenerationConfig.ResponseMimeType != "application/json" {
			t.Fatalf("responseMimeType = %q", req.GenerationConfig.ResponseMimeType)
		}
		if req.SystemInstruction == nil || len(req.SystemInstruction.Parts) == 0 {
			t.Fatal("missing system instruction")
		}
		resp := chatResponse{
			Candidates: []candidate{{
				Content: content{
					Parts: []part{{Text: `{"intent":"find_file","params":{"filename":"*.log"},"confidence":0.9}`}},
				},
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c, err := NewClient("AIza-test", "gemini-2.0-flash", srv.URL, 5*time.Second, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.Chat(context.Background(), "system prompt", "find log files", map[string]any{"type": "object"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Intent != "find_file" || out.Params["filename"] != "*.log" {
		t.Fatalf("out = %+v", out)
	}
	if out.Confidence != 0.9 {
		t.Fatalf("confidence = %v", out.Confidence)
	}
}

func TestClientChatHTTP500(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, _ := NewClient("AIza-test", "m", srv.URL, time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestClientChatHTTP400(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad request"}}`))
	}))
	defer srv.Close()
	c, _ := NewClient("AIza-test", "m", srv.URL, time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrInvalidResp) {
		t.Fatalf("err = %v, want ErrInvalidResp", err)
	}
}

func TestClientChatHTTP429(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":429,"message":"quota exceeded"}}`))
	}))
	defer srv.Close()
	c, _ := NewClient("AIza-test", "m", srv.URL, time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestClientChatEmptyCandidates(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(chatResponse{Candidates: nil})
	}))
	defer srv.Close()
	c, _ := NewClient("AIza-test", "m", srv.URL, time.Second, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, errNoMatch) {
		t.Fatalf("err = %v, want errNoMatch", err)
	}
}

func TestClientChatTimeout(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c, _ := NewClient("AIza-test", "m", srv.URL, 50*time.Millisecond, nil)
	_, err := c.Chat(context.Background(), "s", "u", nil)
	if !errors.Is(err, providers.ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
}

func TestClientExplainChatHappyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := chatResponse{
			Candidates: []candidate{{
				Content: content{
					Parts: []part{{Text: "This command finds log files recursively."}},
				},
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c, _ := NewClient("AIza-test", "m", srv.URL, 5*time.Second, nil)
	text, err := c.ExplainChat(context.Background(), "system", "explain find")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "log files") {
		t.Fatalf("text = %q", text)
	}
}
