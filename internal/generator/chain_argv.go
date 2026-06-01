package generator

import "strings"

// ChainFromArgv splits flat argv on connector tokens (|, &&, ;) into a CommandChain.
func ChainFromArgv(argv []string) *CommandChain {
	if len(argv) < 3 {
		return nil
	}
	type seg struct {
		tokens   []string
		after    ChainConnector
		hasAfter bool
	}
	var segments []seg
	cur := make([]string, 0, len(argv))
	flush := func(after ChainConnector, has bool) {
		if len(cur) == 0 {
			return
		}
		segments = append(segments, seg{tokens: append([]string(nil), cur...), after: after, hasAfter: has})
		cur = cur[:0]
	}
	for _, tok := range argv {
		switch tok {
		case "|":
			flush(ChainPipe, true)
		case "&&":
			flush(ChainAnd, true)
		case ";":
			flush(ChainAnd, true)
		default:
			cur = append(cur, tok)
		}
	}
	flush(ChainPipe, false)
	if len(segments) < 2 {
		return nil
	}
	stages := make([]ChainStage, len(segments))
	conns := make([]ChainConnector, len(segments)-1)
	for i, s := range segments {
		toks := make([]ChainToken, 0, len(s.tokens))
		for _, v := range s.tokens {
			toks = append(toks, ChainToken{Value: v, Expr: tokenLooksLikeExpr(v)})
		}
		stages[i] = ChainStage{Tokens: toks}
		if i < len(segments)-1 {
			if s.hasAfter {
				conns[i] = s.after
			} else {
				conns[i] = ChainPipe
			}
		}
	}
	return &CommandChain{Stages: stages, Connectors: conns}
}

// ArgvHasChainConnector reports whether argv contains a chain connector token.
func ArgvHasChainConnector(argv []string) bool {
	for _, tok := range argv {
		switch tok {
		case "|", "&&", ";":
			return true
		}
	}
	return false
}

func tokenLooksLikeExpr(v string) bool {
	if strings.HasPrefix(strings.TrimSpace(v), "{") {
		return true
	}
	return strings.ContainsRune(v, '$')
}
