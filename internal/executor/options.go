package executor

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/policy"
	"github.com/alibaba40core/clx/internal/risk"
)

var (
	ErrMissingRisk    = errors.New("executor: missing risk assessment")
	ErrMissingPolicy  = errors.New("executor: missing policy result")
	ErrMissingProfile = errors.New("executor: missing system profile")
	ErrEmptyArgv      = errors.New("executor: empty argv")
)

// TimeoutError is returned by Run when the configured execution.timeout elapses
// before the child process exits. It is distinct from a real non-zero exit so
// the CLI can render a useful message instead of "exit status 1".
type TimeoutError struct {
	After time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timed out after %s (configure execution.timeout)", e.After)
}

// RunConfig holds execution parameters.
type RunConfig struct {
	Timeout    time.Duration
	Stdout     io.Writer
	Stderr     io.Writer
	Risk       risk.RiskAssessment
	Policy     policy.Result
	Profile    environment.SystemProfile
	HasRisk    bool
	HasPolicy  bool
	HasProfile bool
}

// Option configures Run.
type Option func(*RunConfig)

// WithTimeout sets the subprocess timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *RunConfig) { c.Timeout = d }
}

// WithIO sets stdout/stderr writers.
func WithIO(stdout, stderr io.Writer) Option {
	return func(c *RunConfig) {
		c.Stdout = stdout
		c.Stderr = stderr
	}
}

// WithRisk attaches the mandatory risk assessment.
func WithRisk(r risk.RiskAssessment) Option {
	return func(c *RunConfig) {
		c.Risk = r
		c.HasRisk = true
	}
}

// WithPolicy attaches the mandatory policy result.
func WithPolicy(p policy.Result) Option {
	return func(c *RunConfig) {
		c.Policy = p
		c.HasPolicy = true
	}
}

// WithProfile supplies the system profile for host resolution (PowerShell vs pwsh).
func WithProfile(p environment.SystemProfile) Option {
	return func(c *RunConfig) {
		c.Profile = p
		c.HasProfile = true
	}
}
