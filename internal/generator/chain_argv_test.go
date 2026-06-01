package generator

import "testing"

func TestChainFromArgvPipe(t *testing.T) {
	t.Parallel()
	c := ChainFromArgv([]string{"ps", "aux", "|", "grep", "node"})
	if c == nil || len(c.Stages) != 2 {
		t.Fatalf("chain=%+v", c)
	}
	if c.Connectors[0] != ChainPipe {
		t.Fatalf("connector=%v", c.Connectors[0])
	}
}
