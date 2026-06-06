package pipeline

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

var dotFrames = []string{".", "..", "..."}

type progress struct {
	w       io.Writer
	width   int
	enabled bool
	mu      sync.Mutex
}

func newProgress(out io.Writer) *progress {
	return &progress{
		w:       out,
		enabled: progressEnabled(out),
	}
}

func progressEnabled(w io.Writer) bool {
	if os.Getenv("CLX_NO_PROGRESS") != "" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isTerminalWriter(w)
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// Spin shows an in-place dot animation while work is in progress.
// stop() clears the line on the same stream used for results.
func (p *progress) Spin() (stop func()) {
	if p == nil || !p.enabled {
		return func() {}
	}
	p.mu.Lock()
	done := make(chan struct{})
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		p.spinLoop(done)
	}()
	p.mu.Unlock()
	return func() {
		close(done)
		<-exited
		p.mu.Lock()
		p.clearLocked()
		p.mu.Unlock()
	}
}

func (p *progress) spinLoop(done <-chan struct{}) {
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()
	frame := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			p.mu.Lock()
			line := dotFrames[frame%len(dotFrames)]
			frame++
			p.writeLine(line)
			p.mu.Unlock()
		}
	}
}

// FinishOutput clears any spinner residue before writing results.
func (p *progress) FinishOutput() {
	if p == nil || !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clearLocked()
}

func (p *progress) clearLocked() {
	if p.width <= 0 {
		return
	}
	pad := strings.Repeat(" ", p.width)
	fmt.Fprintf(p.w, "\r%s\r\n", pad)
	p.width = 0
}

func (p *progress) writeLine(line string) {
	if len(line) > p.width {
		p.width = len(line)
	}
	pad := ""
	if len(line) < p.width {
		pad = strings.Repeat(" ", p.width-len(line))
	}
	fmt.Fprintf(p.w, "\r%s%s", line, pad)
}
