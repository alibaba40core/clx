package parser


// ShellChain is a parsed multi-stage shell command (connectors between stages).
type ShellChain struct {
	Stages     [][]string
	Connectors []string // "pipe" or "and"
}

// SplitShellChain splits tokenized input on |, &&, or ; into stages.
func SplitShellChain(tokens []string) (*ShellChain, bool) {
	if len(tokens) < 3 {
		return nil, false
	}
	type seg struct {
		tokens   []string
		after    string
		hasAfter bool
	}
	var segments []seg
	cur := make([]string, 0, len(tokens))
	flush := func(after string, has bool) {
		if len(cur) == 0 {
			return
		}
		segments = append(segments, seg{tokens: append([]string(nil), cur...), after: after, hasAfter: has})
		cur = cur[:0]
	}
	for _, tok := range tokens {
		switch tok {
		case "|":
			flush("pipe", true)
		case "&&", ";":
			flush("and", true)
		default:
			cur = append(cur, tok)
		}
	}
	flush("pipe", false)
	if len(segments) < 2 {
		return nil, false
	}
	stages := make([][]string, len(segments))
	conns := make([]string, len(segments)-1)
	for i, s := range segments {
		stages[i] = s.tokens
		if i < len(segments)-1 {
			if s.hasAfter {
				conns[i] = s.after
			} else {
				conns[i] = "pipe"
			}
		}
	}
	return &ShellChain{Stages: stages, Connectors: conns}, true
}
