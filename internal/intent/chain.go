package intent

// ChainSpec is a YAML chain strategy (stages + connectors).
type ChainSpec struct {
	Stages     []ChainStageSpec
	Connectors []string // "pipe", "and"
}

// ChainStageSpec is one stage in a chain rule.
type ChainStageSpec struct {
	Tokens []ChainTokenSpec
	Argv   []string // shorthand: plain tokens, expr false
}

// ChainTokenSpec is one token in a chain stage.
type ChainTokenSpec struct {
	Value string
	Expr  bool
}

// HasChain reports whether the strategy defines a multi-stage chain.
func (s Strategy) HasChain() bool {
	return s.Chain != nil && len(s.Chain.Stages) >= 2
}
