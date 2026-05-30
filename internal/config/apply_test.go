package config

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/alibaba40core/clx/internal/yamlutil"
)

func TestApplyWarnsOnLegacyLevelKey(t *testing.T) {
	t.Parallel()
	root, err := yamlutil.Decode(strings.NewReader("safety:\n  level: medium\n"))
	if err != nil {
		t.Fatal(err)
	}
	warnLegacySafetyLevel = sync.Once{}
	var stderr bytes.Buffer
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old; _ = w.Close() }()

	cfg := Default()
	applyNode(&cfg, root)
	_ = w.Close()
	_, _ = stderr.ReadFrom(r)
	msg := stderr.String()
	if !strings.Contains(msg, "safety.level is deprecated") {
		t.Fatalf("stderr = %q", msg)
	}
}

func TestApplyAcceptsLegacyLevelKey(t *testing.T) {
	t.Parallel()
	root, err := yamlutil.Decode(strings.NewReader("safety:\n  level: medium\n"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	applyNode(&cfg, root)
	if cfg.Safety.Mode != "medium" {
		t.Fatalf("Mode = %q, want medium", cfg.Safety.Mode)
	}
}

func TestApplyNormalizesLegacyValues(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{"safe", "low"},
		{"full", "high"},
		{"medium", "medium"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			root, err := yamlutil.Decode(strings.NewReader("safety:\n  level: " + tc.input + "\n"))
			if err != nil {
				t.Fatal(err)
			}
			cfg := Default()
			applyNode(&cfg, root)
			if cfg.Safety.Mode != tc.want {
				t.Fatalf("Mode = %q, want %q", cfg.Safety.Mode, tc.want)
			}
		})
	}
}

func TestApplyPrefersModeOverLegacy(t *testing.T) {
	t.Parallel()
	root, err := yamlutil.Decode(strings.NewReader("safety:\n  mode: high\n  level: safe\n"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	applyNode(&cfg, root)
	if cfg.Safety.Mode != "high" {
		t.Fatalf("Mode = %q, want high", cfg.Safety.Mode)
	}
}

func TestValidateRejectsLegacyValuesAsMode(t *testing.T) {
	t.Parallel()
	c := Default()
	c.Safety.Mode = "full"
	if err := Validate(c); err == nil {
		t.Fatal("expected error for legacy value full used as Mode")
	}
}

func TestEncodeRoundTripIncludesMode(t *testing.T) {
	t.Parallel()
	def := Default()
	var buf bytes.Buffer
	if err := Encode(def, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "mode: medium") {
		t.Fatalf("encoded YAML missing mode: %s", buf.String())
	}
	root, err := yamlutil.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	applyNode(&cfg, root)
	if cfg.Safety.Mode != "medium" {
		t.Fatalf("Mode = %q", cfg.Safety.Mode)
	}
}
