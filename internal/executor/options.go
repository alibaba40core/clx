package executor

import (
	"errors"
	"io"
	"time"

	"github.com/alibaba40core/clx/internal/environment"
	"github.com/alibaba40core/clx/internal/policy"
	"github.com/alibaba40core/clx/internal/risk"
)

var (
	ErrMissingRisk   = errors.New("executor: missing risk assessment")
	ErrMissingPolicy = errors.New("executor: missing policy result")
	ErrEmptyArgv     = errors.New("executor: empty argv")
)

// RunConfig holds execution parameters.
type RunConfig struct {
	Timeout   time.Duration
	Stdout    io.Writer
	Stderr    io.Writer
	Risk      risk.RiskAssessment
	Policy    policy.Result
	Profile   environment.SystemProfile
	HasRisk   bool
	HasPolicy bool
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
