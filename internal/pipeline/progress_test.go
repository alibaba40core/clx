package pipeline

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestProgressDisabledNoOutput(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	p := newProgress(&out)
	if p.enabled {
		t.Fatal("expected progress disabled for non-terminal writer")
	}
	stop := p.Spin()
	time.Sleep(20 * time.Millisecond)
	stop()
	p.FinishOutput()
	if out.Len() != 0 {
		t.Fatalf("disabled progress wrote %q", out.String())
	}
}

func TestProgressStopBeforeDisplayNoOverlap(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	p := &progress{
		w:       &out,
		enabled: true,
	}

	stop := p.Spin()
	time.Sleep(500 * time.Millisecond)
	stop()
	p.FinishOutput()

	fmt.Fprint(&out, "Intent:      ai-generated command\n")
	fmt.Fprint(&out, "Proceed? [Y/n] ")

	got := out.String()
	if strings.Contains(got, "...Intent") || strings.Contains(got, "..Intent") || strings.Contains(got, ".Intent") {
		t.Fatalf("spinner overlapped display output: %q", got)
	}
	if strings.Contains(got, "...Proceed") || strings.Contains(got, "..Proceed") {
		t.Fatalf("spinner overlapped confirm prompt: %q", got)
	}
	if !strings.Contains(got, "Intent:      ai-generated command") {
		t.Fatalf("missing display output: %q", got)
	}
}

func TestProgressFinishOutputClearsLine(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	p := &progress{
		w:       &out,
		enabled: true,
	}
	p.writeLine("...")
	p.FinishOutput()
	if !strings.HasSuffix(out.String(), "\r\n") {
		t.Fatalf("expected cleared line ending with newline, got %q", out.String())
	}
}

func TestProgressSpinStopClears(t *testing.T) {
	var out bytes.Buffer
	p := &progress{
		w:       &out,
		enabled: true,
	}
	stop := p.Spin()
	time.Sleep(500 * time.Millisecond)
	stop()
	p.FinishOutput()
	if !strings.HasSuffix(out.String(), "\r\n") {
		t.Fatalf("expected spinner cleared, got %q", out.String())
	}
}
