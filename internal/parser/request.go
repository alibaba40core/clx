package parser

// InputType classifies how the user input should be interpreted.
type InputType int

const (
	InputShell InputType = iota
	InputNaturalLanguage
	InputPartialShell
	InputCLXInvocation
	InputChainedShell
)

// String returns a stable name for InputType.
func (t InputType) String() string {
	switch t {
	case InputShell:
		return "Shell"
	case InputNaturalLanguage:
		return "NaturalLanguage"
	case InputPartialShell:
		return "PartialShell"
	case InputCLXInvocation:
		return "CLXInvocation"
	case InputChainedShell:
		return "ChainedShell"
	default:
		return "Unknown"
	}
}

// Request is the normalized parser output (see doc/architecture.md §3.2).
type Request struct {
	RawInput       string
	EffectiveInput string // after alias expansion (equals RawInput when unchanged)
	InputType      InputType
	Tokens         []string
	Args           map[string]string
	ShellChain     *ShellChain // set for InputChainedShell
}
