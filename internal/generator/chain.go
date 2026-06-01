package generator

// ChainConnector is a logical join between command stages (CLX maps to shell symbols).
type ChainConnector int

const (
	ChainPipe ChainConnector = iota
	ChainAnd
)

// ChainToken is one argument in a chain stage.
type ChainToken struct {
	Value string
	Expr  bool // scriptblock / predicate; validated separately
}

// ChainStage is one segment of a chained command.
type ChainStage struct {
	Tokens []ChainToken
}

// CommandChain is an ordered list of stages with connectors between them.
type CommandChain struct {
	Stages     []ChainStage
	Connectors []ChainConnector // len == len(Stages)-1
}

// PlainArgv returns argv tokens for a stage.
func (s ChainStage) PlainArgv() []string {
	out := make([]string, 0, len(s.Tokens))
	for _, tok := range s.Tokens {
		if tok.Value != "" {
			out = append(out, tok.Value)
		}
	}
	return out
}

// HasExpr reports whether the stage has an expression token.
func (s ChainStage) HasExpr() bool {
	for _, tok := range s.Tokens {
		if tok.Expr {
			return true
		}
	}
	return false
}

// HasExpr reports whether any stage has an expression token.
func (c *CommandChain) HasExpr() bool {
	if c == nil {
		return false
	}
	for _, st := range c.Stages {
		if st.HasExpr() {
			return true
		}
	}
	return false
}

// NormalizeConnectors fills missing connectors with ChainPipe.
func (c *CommandChain) NormalizeConnectors() {
	if c == nil || len(c.Stages) < 2 {
		return
	}
	need := len(c.Stages) - 1
	if len(c.Connectors) >= need {
		c.Connectors = c.Connectors[:need]
		return
	}
	out := make([]ChainConnector, need)
	for i := 0; i < need; i++ {
		if i < len(c.Connectors) {
			out[i] = c.Connectors[i]
		} else {
			out[i] = ChainPipe
		}
	}
	c.Connectors = out
}
