package pipeline

import (
	"io"
	"log/slog"
	"os"
)

// Options configures pipeline execution.
type Options struct {
	Explain bool
	DryRun  bool
	Yes     bool
	Logger  *slog.Logger
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

// WithDefaults fills nil writers and stdin.
func (o *Options) WithDefaults() {
	if o.Stdin == nil {
		o.Stdin = os.Stdin
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
}
