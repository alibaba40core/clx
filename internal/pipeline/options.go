package pipeline

import (
	"io"
	"os"
)

// Options configures pipeline execution.
type Options struct {
	Explain bool
	DryRun  bool
	Yes     bool
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
