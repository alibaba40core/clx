package providers

import (
	"strings"
	"testing"
)

func TestRedactHTTPBodyStripsSecrets(t *testing.T) {
	t.Parallel()
	body := []byte(`{"error":{"message":"Invalid api_key=sk-secret1234567890abcdef"}}`)
	redacted := RedactHTTPBody(body)
	if strings.Contains(redacted, "sk-secret1234567890abcdef") {
		t.Fatalf("secret leaked: %q", redacted)
	}
}

func TestRedactHTTPBodyCapsLength(t *testing.T) {
	t.Parallel()
	body := make([]byte, 512)
	for i := range body {
		body[i] = 'x'
	}
	redacted := RedactHTTPBody(body)
	if len(redacted) > maxLogSnippet+3 {
		t.Fatalf("len = %d", len(redacted))
	}
	if !strings.HasSuffix(redacted, "…") {
		t.Fatalf("expected truncation suffix: %q", redacted)
	}
}

func TestRedactErrorStripsSecrets(t *testing.T) {
	t.Parallel()
	msg := RedactError(errString("openai failed: Bearer sk-live-abcdefghijklmnopqrstuvwxyz"))
	if strings.Contains(msg, "sk-live-abcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("secret leaked: %q", msg)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
