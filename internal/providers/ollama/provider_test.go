package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/providers"
)

func TestProviderResolveIntentRoundTrip(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Format == nil {
			t.Fatal("expected format schema")
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Message: struct {
				Content string `json:"content"`
			}{Content: `{"intent":"list_dir","params":{},"confidence":0.8}`},
		})
	}))
	defer srv.Close()

	p, err := NewProvider(srv.URL, "qwen3:4b", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	out, err := p.ResolveIntent(context.Background(), providers.IntentRequest{
		RawInput:     "show files here",
		KnownIntents: []string{"list_dir", "find_file"},
		Profile: environment.SystemProfile{
			OS: "linux", Shell: "bash", Terminal: "xterm",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Intent != "list_dir" {
		t.Fatalf("intent = %q", out.Intent)
	}
}
