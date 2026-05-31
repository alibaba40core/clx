package config

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmbeddedShellIntegrationExplainOnly(t *testing.T) {
	t.Parallel()
	body := EmbeddedShellIntegration()
	if !bytes.Contains(body, []byte("--explain")) {
		t.Fatalf("shell integration snippet must reference --explain, got:\n%s", body)
	}
	lower := strings.ToLower(string(body))
	for _, forbidden := range []string{"auto-execute", "auto execute", "-y ", " --yes"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("shell integration must not suggest auto-run (%q)", forbidden)
		}
	}
}
