package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alibaba40core/clx/internal/providers"
)

func TestProviderName(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	p, err := newTestProvider(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "gemini" {
		t.Fatalf("Name() = %q", p.Name())
	}
}

func TestProviderResolveIntent(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := chatResponse{
			Candidates: []candidate{{
				Content: content{
					Parts: []part{{Text: `{"intent":"find_file","params":{"filename":"*.go"},"confidence":0.95}`}},
				},
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := newTestProvider(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := p.ResolveIntent(context.Background(), providers.IntentRequest{
		RawInput:     "find go files",
		KnownIntents: []string{"find_file", "list_dir"},
		RuleParams:   map[string][]string{"find_file": {"filename"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Intent != "find_file" {
		t.Fatalf("Intent = %q", resp.Intent)
	}
	if resp.Params["filename"] != "*.go" {
		t.Fatalf("Params = %v", resp.Params)
	}
}

func TestProviderImplementsCommandGenerator(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := chatResponse{
			Candidates: []candidate{{
				Content: content{
					Parts: []part{{Text: `{"argv":["find",".","-name","*.go"],"shell":"bash","explanation":"find go files","confidence":0.8}`}},
				},
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := newTestProvider(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	var cg providers.CommandGenerator = p
	resp, err := cg.GenerateCommand(context.Background(), providers.CommandRequest{
		RawInput: "find go files",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Argv) != 4 || resp.Argv[0] != "find" {
		t.Fatalf("Argv = %v", resp.Argv)
	}
}

func newTestProvider(baseURL string) (*Provider, error) {
	client, err := NewClient("AIza-test", "gemini-2.0-flash", baseURL, 5*time.Second, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client}, nil
}
